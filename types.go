package awsmocker

import (
	"net"
	"net/http"
)

const DefaultAccountId = "555555555555"
const DefaultRegion = "us-east-1"

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
