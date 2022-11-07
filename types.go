package awsmocker

import (
	"net"
)

const DefaultAccountId = "555555555555"
const DefaultRegion = "us-east-1"

type TestingT interface {
	Setenv(key, value string)
	TempDir() string
	Cleanup(func())
	FailNow()
	Log(args ...any)
	Logf(format string, args ...any)
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
