package awsmocker

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// const (
// 	envAwsCaBundle       = "AWS_CA_BUNDLE"
// 	envAwsAccessKey      = "AWS_ACCESS_KEY_ID"
// 	envAwsSecretKey      = "AWS_SECRET_ACCESS_KEY"
// 	envAwsSessionToken   = "AWS_SESSION_TOKEN"
// 	envAwsEc2MetaDisable = "AWS_EC2_METADATA_DISABLED"
// 	envAwsContCredUri    = "AWS_CONTAINER_CREDENTIALS_FULL_URI"
// 	envAwsContCredRelUri = "AWS_CONTAINER_CREDENTIALS_RELATIVE_URI"
// 	envAwsContAuthToken  = "AWS_CONTAINER_AUTHORIZATION_TOKEN"
// 	envAwsConfigFile     = "AWS_CONFIG_FILE"
// 	envAwsSharedCredFile = "AWS_SHARED_CREDENTIALS_FILE"
// 	envAwsWebIdentTFile  = "AWS_WEB_IDENTITY_TOKEN_FILE"
// 	envAwsDefaultRegion  = "AWS_DEFAULT_REGION"

// 	// AWS_EC2_METADATA_SERVICE_ENDPOINT_MODE
// 	// AWS_EC2_METADATA_SERVICE_ENDPOINT
// )

type mocker struct {
	t          TestingT
	timeout    time.Duration
	httpServer *httptest.Server

	verbose      bool
	debugTraffic bool

	// usingAwsConfig     bool
	doNotOverrideCreds bool
	doNotFailUnhandled bool

	// originalEnv map[string]*string

	mocks []*MockedEndpoint

	// counter used by the middleware to track requests
	mwReqCounter *atomic.Uint64
	requestLog   *sync.Map

	awsConfig aws.Config

	noMiddleware bool
}

func (m *mocker) init() {
	// m.originalEnv = make(map[string]*string, 10)
	m.requestLog = &sync.Map{}
	m.mwReqCounter = &atomic.Uint64{}
}

// Overrides an environment variable and then adds it to the stack to undo later
// func (m *mocker) setEnv(k string, v any) {
// 	return
// 	val, ok := os.LookupEnv(k)
// 	if ok {
// 		m.originalEnv[k] = &val
// 	} else {
// 		m.originalEnv[k] = nil
// 	}

// 	switch nval := v.(type) {
// 	case string:
// 		err := os.Setenv(k, nval)
// 		if err != nil {
// 			m.t.Errorf("Unable to set env var '%s': %s", k, err)
// 		}
// 	case nil:
// 		err := os.Unsetenv(k)
// 		if err != nil {
// 			m.t.Errorf("Unable to unset env var '%s': %s", k, err)
// 		}
// 	default:
// 		panic("WRONG ENV VAR VALUE TYPE: must be nil or a string")
// 	}
// }

// func (m *mocker) revertEnv() {
// 	for k, v := range m.originalEnv {
// 		if v == nil {
// 			_ = os.Unsetenv(k)
// 		} else {
// 			_ = os.Setenv(k, *v)
// 		}
// 	}
// }

func (m *mocker) Start() {
	m.init()

	m.t.Cleanup(m.Shutdown)

	for i := range m.mocks {
		m.mocks[i].prep()
	}

	// ts := httptest.NewServer(m)
	// m.httpServer = ts

	// m.setEnv(envAwsEc2MetaDisable, "true")
	// m.setEnv(envAwsDefaultRegion, DefaultRegion)

	// if !m.doNotOverrideCreds {
	// 	m.setEnv(envAwsAccessKey, "fakekey")
	// 	m.setEnv(envAwsSecretKey, "fakesecret")
	// 	m.setEnv(envAwsSessionToken, "faketoken")
	// 	m.setEnv(envAwsConfigFile, "fakeconffile")
	// 	m.setEnv(envAwsSharedCredFile, "fakesharedfile")
	// }

}

func (m *mocker) Shutdown() {
	if m.httpServer != nil {
		m.httpServer.Close()
	}
	m.requestLog.Clear()

	// m.revertEnv()
}

func (m *mocker) startServer() {
	if m.httpServer != nil {
		return
	}
	m.httpServer = httptest.NewServer(m)
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

func (m *mocker) RoundTrip(req *http.Request) (*http.Response, error) {
	_, resp := m.handleRequest(req)
	return resp, nil
}

func (m *mocker) handleRequest(req *http.Request) (*http.Request, *http.Response) {
	recvReq := newReceivedRequest(req)
	recvReq.mocker = m

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
