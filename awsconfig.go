package awsmocker

import (
	"context"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	// awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
)

// If your application is setup to where you can provide an aws.Config object for your clients,
// then using the one provided by this method will make testing much easier.
func (m *mocker) buildAwsConfig(opts ...AwsLoadOptionsFunc) aws.Config {

	// httpClient := awshttp.NewBuildableClient().WithTimeout(10 * time.Second).WithTransportOptions(func(t *http.Transport) {
	// 	proxyUrl, _ := url.Parse(m.httpServer.URL)
	// 	t.Proxy = http.ProxyURL(proxyUrl)

	// 	// remove the need for CA bundle?
	// 	// t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// })
	// _ = httpClient

	c := &http.Client{
		Transport: m,
		Timeout:   2 * time.Second,
	}

	options := make([]AwsLoadOptionsFunc, 0, 15)

	options = append(options, config.WithDisableRequestCompression(aws.Bool(true)))
	options = append(options, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("XXfakekey", "XXfakesecret", "xxtoken")))
	options = append(options, config.WithDefaultRegion(DefaultRegion))
	// options = append(options, config.WithHTTPClient(httpClient))
	// options = append(options, config.WithHTTPClient(m))
	options = append(options, config.WithHTTPClient(c))
	// options = append(options, config.WithCustomCABundle(bytes.NewReader(caCert)))
	options = append(options, config.WithRetryer(func() aws.Retryer {
		return aws.NopRetryer{}
	}))

	// apply the options the user wanted
	options = append(options, opts...)

	options = append(options, addMiddlewareConfigOption(m))

	cfg, err := config.LoadDefaultConfig(context.TODO(), options...)
	if err != nil {
		panic(err)
	}

	return cfg
}
