package awsmocker_test

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/webdestroya/awsmocker"
)

func TestResponseDebugLogging(t *testing.T) {
	info := awsmocker.Start(t, awsmocker.WithoutDefaultMocks(),
		awsmocker.WithMocks(&awsmocker.MockedEndpoint{
			Request: &awsmocker.MockedRequest{
				Hostname: "httptest.com",
			},
			Response: awsmocker.MockResponse_Error(400, "SomeCode_HTTP", "SomeMessage"),
		}),
		awsmocker.WithMocks(&awsmocker.MockedEndpoint{
			Request: &awsmocker.MockedRequest{
				Hostname: "httpstest.com",
			},
			Response: awsmocker.MockResponse_Error(401, "SomeCode_HTTPS", "SomeMessage"),
		}),
	)

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: func(r *http.Request) (*url.URL, error) {
				return url.Parse(info.ProxyURL())
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	tables := []struct {
		method       string
		url          string
		statusCode   int
		expErrorCode string
	}{

		{http.MethodGet, "http://httptest.com/err400", 400, "SomeCode_HTTP"},
		{http.MethodGet, "https://httpstest.com/err401", 401, "SomeCode_HTTPS"},
	}

	for _, table := range tables {
		req, err := http.NewRequest(table.method, table.url, nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// dump, err := httputil.DumpResponse(resp, true)
		// if err != nil {
		// 	t.Error(err)
		// }
		// fmt.Printf("RECEIVED:\n%s\n", dump)

		require.Equal(t, table.statusCode, resp.StatusCode)

		respBodyRaw, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		respBody := string(respBodyRaw)
		if table.expErrorCode != "" {
			require.Contains(t, respBody, table.expErrorCode)
		}

	}

}
