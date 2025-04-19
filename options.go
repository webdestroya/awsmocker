package awsmocker

import (
	"time"
)

type mockerOptions struct {
	// Add extra logging. This is deprecated, you should just use the AWSMOCKER_DEBUG=1 env var and do a targeted test run
	Verbose bool

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

	// Add mocks for the EC2 Instance Metadata Service
	MockEc2Metadata bool

	// By default, receiving an unmatched request will cause the test to be marked as failed
	// you can pass true to this if you do not want to fail your test when the mocker receives an
	// unmatched request
	DoNotFailUnhandledRequests bool

	AwsConfigOptions []AwsLoadOptionsFunc
}

type MockerOptionFunc = func(*mockerOptions)

func WithoutDefaultMocks() MockerOptionFunc {
	return func(o *mockerOptions) {
		o.SkipDefaultMocks = true
	}
}

// Disables setting credential environment variables
// This is dangerous, because if the proxy were to fail, then your requests may actually
// execute on AWS with real credentials.
func WithoutCredentialProtection() MockerOptionFunc {
	return func(o *mockerOptions) {
		o.DoNotOverrideCreds = true
	}
}

func WithoutFailingUnhandledRequests() MockerOptionFunc {
	return func(o *mockerOptions) {
		o.DoNotFailUnhandledRequests = true
	}
}

func WithMockEC2Metadata(enabled bool) MockerOptionFunc {
	return func(o *mockerOptions) {
		o.MockEc2Metadata = enabled
	}
}

func WithAWSConfigOptions(opts ...AwsLoadOptionsFunc) MockerOptionFunc {
	return func(o *mockerOptions) {
		if o.AwsConfigOptions == nil {
			o.AwsConfigOptions = make([]AwsLoadOptionsFunc, 0, len(opts))
		}
		o.AwsConfigOptions = append(o.AwsConfigOptions, opts...)
	}
}

func WithMocks(mocks ...*MockedEndpoint) MockerOptionFunc {
	return func(o *mockerOptions) {
		if o.Mocks == nil {
			o.Mocks = make([]*MockedEndpoint, 0, len(mocks))
		}
		o.Mocks = append(o.Mocks, mocks...)
	}
}

func WithMock(mock *MockedEndpoint) MockerOptionFunc {
	return WithMocks(mock)
}
