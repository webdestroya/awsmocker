package awsmocker

import (
	"slices"
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

	// Timeout for proxied requests.
	Timeout time.Duration

	// The mocks that will be responded to
	Mocks []*MockedEndpoint

	// Add mocks for the EC2 Instance Metadata Service
	MockEc2Metadata bool

	// By default, receiving an unmatched request will cause the test to be marked as failed
	// you can pass true to this if you do not want to fail your test when the mocker receives an
	// unmatched request
	DoNotFailUnhandledRequests bool

	AwsConfigOptions []AwsLoadOptionsFunc
}

type MockerOptionFunc func(*mockerOptions)

var defaultMocks = []*MockedEndpoint{
	MockStsGetCallerIdentityValid,
}

func newOptions() *mockerOptions {
	return &mockerOptions{
		Timeout: 5 * time.Second,
		Mocks:   slices.Clone(defaultMocks),
	}
}

// Default mocks for GetCallerIdentity and role assumptions will not be provided
func WithoutDefaultMocks() MockerOptionFunc {
	return func(o *mockerOptions) {
		o.Mocks = slices.DeleteFunc(o.Mocks, func(m *MockedEndpoint) bool {
			return slices.Contains(defaultMocks, m)
		})
	}
}

// Disables setting credential environment variables
// This is dangerous, because if the proxy were to fail, then your requests may actually
// execute on AWS with real credentials.
// This means if you do not properly configure the mocker, you could end up making real requests to AWS.
// This is not recommended.
// Deprecated: You should really not be using this
func WithoutCredentialProtection() MockerOptionFunc {
	return func(o *mockerOptions) {
		o.DoNotOverrideCreds = true
	}
}

// By default, receiving an unmatched request will cause the test to be marked as failed
// you can pass true to this if you do not want to fail your test when the mocker receives an
// unmatched request
func WithoutFailingUnhandledRequests() MockerOptionFunc {
	return func(o *mockerOptions) {
		o.DoNotFailUnhandledRequests = true
	}
}

// Add mocks for the EC2 Instance Metadata Service
// These are not exhaustive, so if you have a special need you will have to add it.
func WithEC2Metadata(opts ...IMDSMockOptionFunc) MockerOptionFunc {
	return WithMocks(Mock_IMDS_Common(opts...)...)
}

// Additional AWS LoadConfig options to pass along to the Config builder
// Use this if you need to add custom middleware.
// Don't use this to set credentials, or HTTP client, as those will be overridden
func WithAWSConfigOptions(opts ...AwsLoadOptionsFunc) MockerOptionFunc {
	return func(o *mockerOptions) {
		o.AwsConfigOptions = append(o.AwsConfigOptions, opts...)
	}
}

// The mocks that will be responded to
func WithMocks(mocks ...*MockedEndpoint) MockerOptionFunc {
	return func(o *mockerOptions) {
		o.Mocks = append(o.Mocks, mocks...)
	}
}

// If provided, then requests that run longer than this will be terminated.
// Generally you should not need to set this
func WithTimeout(value time.Duration) MockerOptionFunc {
	return func(mo *mockerOptions) {
		mo.Timeout = value
	}
}

// Add extra logging.
//
// Deprecated: you should just use the AWSMOCKER_DEBUG=1 env var and do a targeted test run
func WithVerbosity(value bool) MockerOptionFunc {
	return func(mo *mockerOptions) {
		mo.Verbose = value
	}
}
