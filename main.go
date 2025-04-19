package awsmocker

import (
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// Returned when you start the server, provides you some information if needed
type MockerInfo struct {
	// URL of the proxy server
	ProxyURL string

	// Aws configuration to use
	// Deprecated: use [Config] instead
	awsConfig *aws.Config
}

func (m MockerInfo) Config() aws.Config {
	if m.awsConfig == nil {
		panic("aws config was not setup properly")
	}
	return *m.awsConfig
}

// Use this for custom proxy configurations
func (m MockerInfo) Proxy() func(*http.Request) (*url.URL, error) {
	uri, err := url.Parse(m.ProxyURL)
	return func(r *http.Request) (*url.URL, error) {
		return uri, err
	}
}

func Start(t TestingT, optFns ...MockerOptionFunc) *MockerInfo {

	if h, ok := t.(tHelper); ok {
		h.Helper()
	}

	options := &mockerOptions{
		ReturnAwsConfig:  true,
		Timeout:          5 * time.Second,
		AwsConfigOptions: make([]AwsLoadOptionsFunc, 0, 10),
		Mocks:            make([]*MockedEndpoint, 0, 10),
	}

	for _, optFn := range optFns {

		// makes transitioning somewhat easier
		if optFn == nil {
			continue
		}
		optFn(options)
	}

	options.ReturnAwsConfig = true

	// if options.Timeout == 0 {
	// 	options.Timeout = 5 * time.Second
	// }

	if !options.SkipDefaultMocks {
		options.Mocks = append(options.Mocks, MockStsGetCallerIdentityValid)
	}

	if options.MockEc2Metadata {
		options.Mocks = append(options.Mocks, Mock_IMDS_Common()...)
	}

	// proxy bypass configuration
	// if options.DoNotProxy != "" {
	// 	noProxyStr := os.Getenv("NO_PROXY")
	// 	if noProxyStr == "" {
	// 		noProxyStr = os.Getenv("no_proxy")
	// 	}
	// 	if noProxyStr != "" {
	// 		noProxyStr += ","
	// 	}
	// 	noProxyStr += options.DoNotProxy

	// 	t.Setenv("NO_PROXY", noProxyStr)
	// 	t.Setenv("no_proxy", noProxyStr)
	// }

	mocks := make([]*MockedEndpoint, 0, len(options.Mocks))
	for i := range options.Mocks {
		if options.Mocks[i] == nil {
			continue
		}
		mocks = append(mocks, options.Mocks[i])
	}

	server := &mocker{
		t:                  t,
		timeout:            options.Timeout,
		verbose:            options.Verbose,
		debugTraffic:       getDebugMode(), // options.DebugTraffic,
		doNotOverrideCreds: options.DoNotOverrideCreds,
		doNotFailUnhandled: options.DoNotFailUnhandledRequests,
		mocks:              mocks,
		usingAwsConfig:     true,
	}
	server.Start()

	cfg := server.buildAwsConfig()

	info := &MockerInfo{
		ProxyURL:  server.httpServer.URL,
		awsConfig: &cfg,
	}

	return info
}
