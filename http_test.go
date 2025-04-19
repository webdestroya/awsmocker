package awsmocker_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/webdestroya/awsmocker"
)

func TestProxyHttp(t *testing.T) {
	info := awsmocker.Start(t, awsmocker.WithoutFailingUnhandledRequests())

	transport := http.Transport{
		Proxy: info.Proxy(),
	}
	// transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //set ssl

	client := &http.Client{
		Transport: &transport,
	}
	httpresp, err := client.Get("http://example.com/")
	require.NoError(t, err)
	defer httpresp.Body.Close()
}
