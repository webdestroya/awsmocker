package awsmocker

import (
	"net"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
)

const (
	DefaultAccountId = "555555555555"
	DefaultRegion    = "us-east-1"
	// DefaultInstanceId = "i-000deadbeef"
)

type TestingT interface {
	Setenv(key, value string)
	TempDir() string
	Cleanup(func())
	Fail()
	Errorf(format string, args ...any)
	Logf(format string, args ...any)

	// These must be called from the test goroutine (which we will not be in)
	// so do not use them
	// FailNow()
	// Fatalf(format string, args ...any)
}

type tHelper interface {
	Helper()
}

type halfClosable interface {
	net.Conn
	CloseWrite() error
	CloseRead() error
}

var _ halfClosable = (*net.TCPConn)(nil)

type ResponseEncoding int

const (
	// Default will try to determine encoding via the request headers
	ResponseEncodingDefault ResponseEncoding = iota
	ResponseEncodingJSON
	ResponseEncodingXML
	ResponseEncodingText
)

type MockedRequestHandler = func(*ReceivedRequest) *http.Response

// Come on Amazon...
// Can't use {config.LoadOptionsFunc} because that is not a param for LoadDefaultConfig
// Why would you define the type and then not even use it...
type AwsLoadOptionsFunc = func(*config.LoadOptions) error
