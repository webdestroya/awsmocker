package awsmocker

import (
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func Start(t TestingT, options *MockerOptions) *aws.Config {

	if h, ok := t.(tHelper); ok {
		h.Helper()
	}

	if options == nil {
		options = &MockerOptions{}
	}

	if options.Timeout == 0 {
		options.Timeout = 5 * time.Second
	}

	if !options.SkipDefaultMocks {
		options.Mocks = append(options.Mocks, MockStsGetCallerIdentityValid)
	}

	// proxy bypass configuration
	if options.DoNotProxy != "" {
		noProxyStr := os.Getenv("NO_PROXY")
		if noProxyStr == "" {
			noProxyStr = os.Getenv("no_proxy")
		}
		if noProxyStr != "" {
			noProxyStr += ","
		}
		noProxyStr += options.DoNotProxy

		t.Setenv("NO_PROXY", noProxyStr)
		t.Setenv("no_proxy", noProxyStr)
	}

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
		mocks:              mocks,
		usingAwsConfig:     options.ReturnAwsConfig,
	}
	server.Start()

	if options.ReturnAwsConfig {
		cfg := server.buildAwsConfig()
		return &cfg
	}

	return nil
}
