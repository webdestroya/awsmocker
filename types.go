package awsmocker

import (
	"testing"
)

const DefaultAccountId = "555555555555"

type MockerOptions struct {
	// used to get the TempDir value
	T *testing.T

	// provide a path to a temporary directory used to write the CA Bundle
	TempDir string

	// dump request/responses to the log
	Verbose bool

	// if true, then env vars for various aws credentials will not be set.
	// This is dangerous, because if the proxy were to fail, then your requests may actually
	// execute on AWS with real credentials.
	//
	DoNotOverrideCreds bool

	// if this is true, then default mocks for GetCallerIdentity and role assumptions will not be provided
	SkipDefaultMocks bool

	// The mocks that will be responded to
	Mocks []*MockedEndpoint
}

type ResponseEncoding int

const (
	// Default will try to determine encoding via the request headers
	ResponseEncodingDefault ResponseEncoding = iota
	ResponseEncodingJSON
	ResponseEncodingXML
	ResponseEncodingText
)
