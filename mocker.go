package awsmocker

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	_ "unsafe"

	"github.com/elazarl/goproxy"
)

var GlobalDebugMode = false

// GO actually caches proxy env vars which totally breaks our test flow
// so this hacks in a call to Go's internal method... This is pretty janky

//go:linkname resetProxyConfig net/http.resetProxyConfig
func resetProxyConfig()

func init() {
	resetProxyConfig()
}

const (
	envHttpProxy  = "HTTP_PROXY"
	envHttpsProxy = "HTTPS_PROXY"

	envAwsCaBundle       = "AWS_CA_BUNDLE"
	envAwsAccessKey      = "AWS_ACCESS_KEY_ID"
	envAwsSecretKey      = "AWS_SECRET_ACCESS_KEY"
	envAwsSessionToken   = "AWS_SESSION_TOKEN"
	envAwsEc2MetaDisable = "AWS_EC2_METADATA_DISABLED"
	envAwsContCredUri    = "AWS_CONTAINER_CREDENTIALS_FULL_URI"
	envAwsContAuthToken  = "AWS_CONTAINER_AUTHORIZATION_TOKEN"
	envAwsConfigFile     = "AWS_CONFIG_FILE"
	envAwsSharedCredFile = "AWS_SHARED_CREDENTIALS_FILE"
	envAwsDefaultRegion  = "AWS_DEFAULT_REGION"
)

var relevantEnvVars = [...]string{
	envAwsCaBundle,
	envHttpProxy,
	envHttpsProxy,
	envAwsAccessKey,
	envAwsSecretKey,
	envAwsSessionToken,
	envAwsEc2MetaDisable,
	envAwsContCredUri,
	envAwsContAuthToken,
	envAwsConfigFile,
	envAwsSharedCredFile,
}

type mockerServer struct {
	httpServer      *httptest.Server
	originalEnvVars map[string]*string

	verbose bool

	// GoProxy originals
	origGPCa              tls.Certificate
	origGPOkConnect       *goproxy.ConnectAction
	origGPMitmConnect     *goproxy.ConnectAction
	origGPHttpMitmConnect *goproxy.ConnectAction
	origGPRejectConnect   *goproxy.ConnectAction
}

func (ms *mockerServer) init() {
	ms.originalEnvVars = make(map[string]*string, 10)
}

func (ms *mockerServer) OverrideEnvVar(k string, v interface{}) {
	val, ok := os.LookupEnv(k)
	if ok {
		ms.originalEnvVars[k] = &val
	} else {
		ms.originalEnvVars[k] = nil
	}

	switch nval := v.(type) {
	case string:
		err := os.Setenv(k, nval)
		if err != nil {
			fmt.Printf("SETENV_ERROR(%s):%s\n", k, err)
		}
	case nil:
		err := os.Unsetenv(k)
		if err != nil {
			fmt.Printf("UNSETENV_ERROR(%s):%s\n", k, err)
		}
	default:
		panic("WRONG ENV VAR VALUE TYPE: must be nil or a string")
	}

}

func StartMockServer(options *MockerOptions) (func(), string, error) {
	server := newMockServer(options)
	return server.Close, server.httpServer.URL, nil
}

func newMockServer(options *MockerOptions) *mockerServer {

	if GlobalDebugMode {
		fmt.Println("STARTING MOCK SERVER")
	}

	if options.TempDir == "" && options.T != nil {
		options.TempDir = options.T.TempDir()
	}

	if options.TempDir == "" {
		panic("You must provide T or TempDir")
	}

	if !options.SkipDefaultMocks {
		options.Mocks = append(options.Mocks, MockStsGetCallerIdentityValid)
	}

	// prepare the mock endpoints
	for i := range options.Mocks {
		options.Mocks[i].prep()
	}

	caBundlePath := path.Join(options.TempDir, "awsmockcabundle.pem")

	writeCABundle(caBundlePath)

	server := &mockerServer{
		verbose: options.Verbose,
	}
	server.init()

	server.origGPCa = goproxy.GoproxyCa
	server.origGPOkConnect = goproxy.OkConnect
	server.origGPMitmConnect = goproxy.MitmConnect
	server.origGPHttpMitmConnect = goproxy.HTTPMitmConnect
	server.origGPRejectConnect = goproxy.RejectConnect

	tlsConfig := goproxy.TLSConfigFromCA(&caKeyPair)
	goproxy.GoproxyCa = caKeyPair
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: tlsConfig}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: tlsConfig}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: tlsConfig}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: tlsConfig}

	proxy := goproxy.NewProxyHttpServer()
	// proxy.Verbose = options.Verbose
	proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		return handleRequest(options, req, ctx)
	})

	ts := httptest.NewServer(proxy)
	server.httpServer = ts

	server.OverrideEnvVar(envAwsCaBundle, caBundlePath)
	server.OverrideEnvVar(envHttpProxy, ts.URL)
	server.OverrideEnvVar(envHttpsProxy, ts.URL)
	server.OverrideEnvVar(envAwsEc2MetaDisable, "true")
	server.OverrideEnvVar(envAwsDefaultRegion, "us-east-1")

	if !options.DoNotOverrideCreds {
		server.OverrideEnvVar(envAwsAccessKey, "fakekey")
		server.OverrideEnvVar(envAwsSecretKey, "fakesecret")
		server.OverrideEnvVar(envAwsSessionToken, "faketoken")
		server.OverrideEnvVar(envAwsConfigFile, "fakeconffile")
		server.OverrideEnvVar(envAwsSharedCredFile, "fakesharedfile")
	}

	return server
}

func (m *mockerServer) Close() {

	if GlobalDebugMode {
		fmt.Println("AWSMOCKER: Closing...")
	}

	// reset the env
	for k, v := range m.originalEnvVars {
		if v == nil {
			os.Unsetenv(k)
		} else {
			os.Setenv(k, *v)
		}
	}

	// close the server
	m.httpServer.Close()

	// reset goproxy
	goproxy.GoproxyCa = m.origGPCa
	goproxy.OkConnect = m.origGPOkConnect
	goproxy.MitmConnect = m.origGPMitmConnect
	goproxy.HTTPMitmConnect = m.origGPHttpMitmConnect
	goproxy.RejectConnect = m.origGPRejectConnect

	// reset Go's proxy cache
	resetProxyConfig()
}

func handleRequest(options *MockerOptions, req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	recvReq := newReceivedRequest(req)

	if recvReq.invalid {
		log.Printf("WARN: BAD REQUEST")
		recvReq.DebugDump()
		return req, generateErrorStruct("AccessDenied", "You provided a bad or invalid request").getResponse(recvReq).toHttpResponse(req)
	}

	if GlobalDebugMode {
		recvReq.DebugDump()
	}

	for _, mockEndpoint := range options.Mocks {
		if mockEndpoint.matchRequest(recvReq) {
			return req, mockEndpoint.getResponse(recvReq).toHttpResponse(req)
		}
	}

	log.Printf("WARN: No matching mocks found for this request!")
	if !GlobalDebugMode {
		recvReq.DebugDump()
	}

	return req, generateErrorStruct("AccessDenied", "No matching request mock was found for this").getResponse(recvReq).toHttpResponse(req)
}
