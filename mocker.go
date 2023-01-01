package awsmocker

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"time"
)

const (
	envAwsCaBundle       = "AWS_CA_BUNDLE"
	envAwsAccessKey      = "AWS_ACCESS_KEY_ID"
	envAwsSecretKey      = "AWS_SECRET_ACCESS_KEY"
	envAwsSessionToken   = "AWS_SESSION_TOKEN"
	envAwsEc2MetaDisable = "AWS_EC2_METADATA_DISABLED"
	envAwsContCredUri    = "AWS_CONTAINER_CREDENTIALS_FULL_URI"
	envAwsContCredRelUri = "AWS_CONTAINER_CREDENTIALS_RELATIVE_URI"
	envAwsContAuthToken  = "AWS_CONTAINER_AUTHORIZATION_TOKEN"
	envAwsConfigFile     = "AWS_CONFIG_FILE"
	envAwsSharedCredFile = "AWS_SHARED_CREDENTIALS_FILE"
	envAwsWebIdentTFile  = "AWS_WEB_IDENTITY_TOKEN_FILE"
	envAwsDefaultRegion  = "AWS_DEFAULT_REGION"

	// AWS_EC2_METADATA_SERVICE_ENDPOINT_MODE
	// AWS_EC2_METADATA_SERVICE_ENDPOINT
)

type mocker struct {
	t          TestingT
	timeout    time.Duration
	httpServer *httptest.Server

	verbose      bool
	debugTraffic bool

	usingAwsConfig     bool
	doNotOverrideCreds bool
	doNotFailUnhandled bool

	originalEnv map[string]*string

	mocks []*MockedEndpoint
}

func (m *mocker) init() {
	m.originalEnv = make(map[string]*string, 10)
}

// Overrides an environment variable and then adds it to the stack to undo later
func (m *mocker) setEnv(k string, v any) {
	val, ok := os.LookupEnv(k)
	if ok {
		m.originalEnv[k] = &val
	} else {
		m.originalEnv[k] = nil
	}

	switch nval := v.(type) {
	case string:
		err := os.Setenv(k, nval)
		if err != nil {
			m.t.Errorf("Unable to set env var '%s': %s", k, err)
		}
	case nil:
		err := os.Unsetenv(k)
		if err != nil {
			m.t.Errorf("Unable to unset env var '%s': %s", k, err)
		}
	default:
		panic("WRONG ENV VAR VALUE TYPE: must be nil or a string")
	}
}

func (m *mocker) revertEnv() {
	for k, v := range m.originalEnv {
		if v == nil {
			_ = os.Unsetenv(k)
		} else {
			_ = os.Setenv(k, *v)
		}
	}
}

func (m *mocker) Start() {
	// reset Go's proxy cache
	resetProxyConfig()

	m.init()

	m.t.Cleanup(m.Shutdown)

	for i := range m.mocks {
		m.mocks[i].prep()
	}

	// if we are using aws config, then we don't need this
	if !m.usingAwsConfig {
		caBundlePath := path.Join(m.t.TempDir(), "awsmockcabundle.pem")
		err := writeCABundle(caBundlePath)
		if err != nil {
			m.t.Errorf("Failed to write CA Bundle: %s", err)
		}
		m.setEnv(envAwsCaBundle, caBundlePath)
	}

	ts := httptest.NewServer(m)
	m.httpServer = ts

	m.setEnv("HTTP_PROXY", ts.URL)
	m.setEnv("http_proxy", ts.URL)
	m.setEnv("HTTPS_PROXY", ts.URL)
	m.setEnv("https_proxy", ts.URL)

	// m.setEnv(envAwsEc2MetaDisable, "true")
	m.setEnv(envAwsDefaultRegion, DefaultRegion)

	if !m.doNotOverrideCreds {
		m.setEnv(envAwsAccessKey, "fakekey")
		m.setEnv(envAwsSecretKey, "fakesecret")
		m.setEnv(envAwsSessionToken, "faketoken")
		m.setEnv(envAwsConfigFile, "fakeconffile")
		m.setEnv(envAwsSharedCredFile, "fakesharedfile")
	}

}

func (m *mocker) Shutdown() {
	m.httpServer.Close()

	m.revertEnv()

	// reset Go's proxy cache
	if !m.usingAwsConfig {
		resetProxyConfig()
	}
}

func (m *mocker) Logf(format string, args ...any) {
	if !m.verbose {
		return
	}
	m.printf(format, args...)
}
func (m *mocker) Warnf(format string, args ...any) {
	m.printf("WARN: "+format, args...)
}

func (m *mocker) printf(format string, args ...any) {
	m.t.Logf("[AWSMOCKER] "+format, args...)
}

func (m *mocker) handleRequest(req *http.Request) (*http.Request, *http.Response) {
	recvReq := newReceivedRequest(req)

	// if recvReq.invalid {
	// 	recvReq.DebugDump()
	// 	m.t.Errorf("You provided an invalid request")
	// 	return req, generateErrorStruct(http.StatusNotImplemented, "AccessDenied", "You provided a bad or invalid request").getResponse(recvReq).toHttpResponse(req)
	// }

	if m.debugTraffic {
		recvReq.DebugDump()
	}

	for _, mockEndpoint := range m.mocks {
		if mockEndpoint.matchRequest(recvReq) {
			// increment it's matcher count
			mockEndpoint.Request.incMatchCount()

			// build the response
			return req, mockEndpoint.getResponse(recvReq).toHttpResponse(req)
		}
	}

	if !m.doNotFailUnhandled {
		m.t.Errorf("No matching request mock was found for this request: %s", recvReq.Inspect())
	}

	return req, generateErrorStruct(http.StatusNotImplemented, "AccessDenied", "No matching request mock was found for this").getResponse(recvReq).toHttpResponse(req)
}

func (m *mocker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hostname := r.URL.Hostname()

	if m.verbose {
		buf := new(bytes.Buffer)
		fmt.Fprintln(buf, "AWSMocker Proxy Request:")
		fmt.Fprintf(buf, "%s %s [%s]\n", r.Method, r.RequestURI, r.Proto)
		fmt.Fprintf(buf, "Host: %s --- Raw: %s\n", hostname, r.Host)
		for k, vlist := range r.Header {
			for _, v := range vlist {
				fmt.Fprintf(buf, "%s: %s\n", k, v)
			}
		}
		m.Logf(buf.String())
	}

	if r.Method == "CONNECT" {
		m.handleHttps(w, r)
		return
	}

	if !r.URL.IsAbs() {
		handleNonProxyRequest.ServeHTTP(w, r)
		return
	}

	// Must be an HTTP call
	m.handleHttp(w, r)

}
