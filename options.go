package awsmocker

import (
	"time"
)

type MockerOptions struct {
	// Add extra logging
	Verbose bool

	// dump request/responses to the log
	// DebugTraffic bool

	// if true, then env vars for various aws credentials will not be set.
	// This is dangerous, because if the proxy were to fail, then your requests may actually
	// execute on AWS with real credentials.
	//
	DoNotOverrideCreds bool

	// if this is true, then default mocks for GetCallerIdentity and role assumptions will not be provided
	SkipDefaultMocks bool

	// WARNING: Setting this to true assumes that you are able to use the config value returned
	// If you do not use the provided config and set this true, then requests will not be routed properly.
	ReturnAwsConfig bool

	// Timeout for proxied requests.
	Timeout time.Duration

	// The mocks that will be responded to
	Mocks []*MockedEndpoint

	// Comma separated list of hostname globs that should not be proxied
	// if you are doing other HTTP/HTTPS requests within your test, you should
	// add the hostnames used to this.
	DoNotProxy string

	// By default, receiving an unmatched request will cause the test to be marked as failed
	// you can pass true to this if you do not want to fail your test when the mocker receives an
	// unmatched request
	// DoNotFailUnhandledRequests bool
}
