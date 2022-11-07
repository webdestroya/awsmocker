package awsmocker_test

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/webdestroya/awsmocker"
)

func TestProxyHttp(t *testing.T) {
	awsmocker.Start(t, nil)

	transport := http.Transport{}
	proxyUrl, _ := url.Parse(os.Getenv("HTTP_PROXY"))
	transport.Proxy = http.ProxyURL(proxyUrl) // set proxy
	// transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //set ssl

	client := &http.Client{
		Transport: &transport,
	}
	httpresp, err := client.Get("http://example.com/")
	require.NoError(t, err)

	data, err := io.ReadAll(httpresp.Body)
	require.NoError(t, err)

	fmt.Println("HTTP GET RESPONSE: ", string(data))
}
