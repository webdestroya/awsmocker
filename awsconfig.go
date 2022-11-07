package awsmocker

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

// If your application is setup to where you can provide an aws.Config object for your clients,
// then using the one provided by this method will make testing much easier.
func (m *mocker) buildAwsConfig() aws.Config {

	httpClient := awshttp.NewBuildableClient().WithTimeout(2 * time.Second).WithTransportOptions(func(t *http.Transport) {
		proxyUrl, _ := url.Parse(m.httpServer.URL)
		t.Proxy = http.ProxyURL(proxyUrl)

		// remove the need for CA bundle?
		t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("XXfakekey", "XXfakesecret", "xxtoken")),
		config.WithDefaultRegion("us-east-1"),
		config.WithHTTPClient(httpClient),
		config.WithRetryer(func() aws.Retryer {
			return aws.NopRetryer{}
		}),
	)
	if err != nil {
		panic(err)
	}

	return cfg
}
