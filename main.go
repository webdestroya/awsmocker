package awsmocker

func Start(t TestingT, optFns ...MockerOptionFunc) *MockerInfo {

	if h, ok := t.(tHelper); ok {
		h.Helper()
	}

	options := newOptions()

	for _, optFn := range optFns {

		// makes transitioning somewhat easier
		if optFn == nil {
			continue
		}
		optFn(options)
	}

	// if options.Timeout == 0 {
	// 	options.Timeout = 5 * time.Second
	// }

	// if !options.SkipDefaultMocks {
	// 	options.Mocks = append(options.Mocks, MockStsGetCallerIdentityValid)
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

	cfg := server.buildAwsConfig(options.AwsConfigOptions...)

	info := &MockerInfo{
		ProxyURL:  server.httpServer.URL,
		awsConfig: &cfg,
	}

	return info
}
