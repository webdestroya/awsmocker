package awsmocker

import (
	"net/http"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
)

// Returned when you start the server, provides you some information if needed
type MockerInfo interface {
	// URL of the proxy server
	ProxyURL() string

	Proxy() func(*http.Request) (*url.URL, error)

	IMDSClient() *imds.Client

	// Aws configuration to use
	Config() aws.Config
}

var _ MockerInfo = (*mocker)(nil)

func (m mocker) Config() aws.Config {
	return m.awsConfig
}

// Use this for custom proxy configurations
func (m mocker) Proxy() func(*http.Request) (*url.URL, error) {
	uri, err := url.Parse(m.ProxyURL())
	return func(_ *http.Request) (*url.URL, error) {
		return uri, err
	}
}

func (m *mocker) ProxyURL() string {
	return m.httpServer.URL
}

func (m mocker) IMDSClient() *imds.Client {
	return imds.NewFromConfig(m.Config())
}
