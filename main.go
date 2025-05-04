package awsmocker

// Start the mocker
func Start(t TestingT, optFns ...MockerOptionFunc) MockerInfo {

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
		noMiddleware:       options.noMiddleware,
		mocks:              mocks,
		usingAwsConfig:     true,
	}
	server.Start()

	server.awsConfig = server.buildAwsConfig(options.AwsConfigOptions...)

	return server
}
