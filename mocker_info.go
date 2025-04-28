package awsmocker

import (
	"net/http"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
)

// Returned when you start the server, provides you some information if needed
type MockerInfo struct {
	// URL of the proxy server
	ProxyURL string

	// Aws configuration to use
	// Deprecated: use [Config] instead
	awsConfig *aws.Config
}

func (m MockerInfo) Config() aws.Config {
	if m.awsConfig == nil {
		panic("aws config was not setup properly")
	}
	return *m.awsConfig
}

// Use this for custom proxy configurations
func (m MockerInfo) Proxy() func(*http.Request) (*url.URL, error) {
	uri, err := url.Parse(m.ProxyURL)
	return func(_ *http.Request) (*url.URL, error) {
		return uri, err
	}
}

func (m MockerInfo) IMDSClient() *imds.Client {
	return imds.NewFromConfig(m.Config())
}
