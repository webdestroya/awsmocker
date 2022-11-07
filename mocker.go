package awsmocker

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"time"
)

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

type mocker struct {
	t          TestingT
	timeout    time.Duration
	httpServer *httptest.Server

	verbose      bool
	debugTraffic bool

	usingAwsConfig     bool
	doNotOverrideCreds bool

	originalEnvVars []string

	mocks []*MockedEndpoint
}

func (m *mocker) Setenv(k, v string) {
	if v == "" {
		_ = os.Unsetenv(k)
	} else {
		_ = os.Setenv(k, v)
	}
}

func (m *mocker) Start() {
	// reset Go's proxy cache
	resetProxyConfig()

	m.originalEnvVars = os.Environ()

	// fmt.Println("ENVVARS", m.originalEnvVars)

	m.t.Cleanup(m.Shutdown)

	for i := range m.mocks {
		m.mocks[i].prep()
	}

	caBundlePath := path.Join(m.t.TempDir(), "awsmockcabundle.pem")
	err := writeCABundle(caBundlePath)
	if err != nil {
		m.Warnf("Failed to write CA Bundle: %s", err)
		m.t.FailNow()
	}
	m.Setenv(envAwsCaBundle, caBundlePath)

	ts := httptest.NewServer(m)
	m.httpServer = ts

	m.Setenv("HTTP_PROXY", ts.URL)
	m.Setenv("http_proxy", ts.URL)
	m.Setenv("HTTPS_PROXY", ts.URL)
	m.Setenv("https_proxy", ts.URL)

	m.Setenv(envAwsEc2MetaDisable, "true")
	m.Setenv(envAwsDefaultRegion, DefaultRegion)

	if !m.doNotOverrideCreds {
		m.Setenv(envAwsAccessKey, "fakekey")
		m.Setenv(envAwsSecretKey, "fakesecret")
		m.Setenv(envAwsSessionToken, "faketoken")
		m.Setenv(envAwsConfigFile, "fakeconffile")
		m.Setenv(envAwsSharedCredFile, "fakesharedfile")
	}

}

func (m *mocker) Shutdown() {
	m.httpServer.Close()

	os.Clearenv()
	for _, e := range m.originalEnvVars {
		pair := strings.SplitN(e, "=", 2)
		_ = os.Setenv(pair[0], pair[1])
	}

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

	if recvReq.invalid {
		recvReq.DebugDump()
		m.t.FailNow()
		return req, generateErrorStruct("AccessDenied", "You provided a bad or invalid request").getResponse(recvReq).toHttpResponse(req)
	}

	if m.debugTraffic {
		recvReq.DebugDump()
	}

	for _, mockEndpoint := range m.mocks {
		if mockEndpoint.matchRequest(recvReq) {
			return req, mockEndpoint.getResponse(recvReq).toHttpResponse(req)
		}
	}

	m.Warnf("WARN: No matching mocks found for this request!")
	if !m.debugTraffic {
		recvReq.DebugDump()
	}

	return req, generateErrorStruct("AccessDenied", "No matching request mock was found for this").getResponse(recvReq).toHttpResponse(req)
}

func (m *mocker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hostname := r.URL.Hostname()

	if m.verbose {
		fmt.Println("AWSMocker Proxy Request:")
		fmt.Printf("%s %s [%s]\n", r.Method, r.RequestURI, r.Proto)
		fmt.Printf("Host: %s --- Raw: %s\n", hostname, r.Host)
		for k, vlist := range r.Header {
			for _, v := range vlist {
				fmt.Printf("%s: %s\n", k, v)
			}
		}
		fmt.Println("---------------------------")
		fmt.Println()
	}

	if r.Method == "CONNECT" {
		m.handleHttps(w, r)
		return
	}

	// Must be an HTTP call
	m.handleHttp(w, r)

}
