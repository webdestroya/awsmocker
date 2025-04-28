package awsmocker_test

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/stretchr/testify/require"
	"github.com/webdestroya/awsmocker"
)

func TestMockResponse_Error(t *testing.T) {
	info := awsmocker.Start(t,
		awsmocker.WithoutDefaultMocks(),
		awsmocker.WithMocks([]*awsmocker.MockedEndpoint{
			{
				Request: &awsmocker.MockedRequest{
					Hostname: "test.com",
					Path:     "/err400",
				},
				Response: awsmocker.MockResponse_Error(400, "SomeCode", "SomeMessage"),
			},
			{
				Request: &awsmocker.MockedRequest{
					Hostname: "test.com",
					Path:     "/err401",
				},
				Response: awsmocker.MockResponse_Error(401, "SomeCode2", "SomeMessage"),
			},
			{
				Request: &awsmocker.MockedRequest{
					Hostname: "test.com",
					Path:     "/err0",
				},
				Response: awsmocker.MockResponse_Error(0, "SomeCode0", "SomeMessage"),
			},
		}...),
	)

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: func(r *http.Request) (*url.URL, error) {
				return url.Parse(info.ProxyURL)
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

		{http.MethodGet, "https://test.com/err400", 400, "SomeCode"},
		{http.MethodGet, "https://test.com/err401", 401, "SomeCode2"},
		{http.MethodGet, "https://test.com/err0", 400, "SomeCode0"},
	}

	for _, table := range tables {
		req, err := http.NewRequest(table.method, table.url, nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, table.statusCode, resp.StatusCode)

		respBodyRaw, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		respBody := string(respBodyRaw)
		if table.expErrorCode != "" {
			require.Contains(t, respBody, table.expErrorCode)
		}
	}

}

func TestMock_Failure(t *testing.T) {
	info := awsmocker.Start(t,
		awsmocker.WithoutDefaultMocks(),
		awsmocker.WithMocks(awsmocker.Mock_Failure("ecs", "ListClusters")),
		awsmocker.WithMocks(awsmocker.Mock_Failure_WithCode(403, "ecs", "ListServices", "SomeCode", "SomeMessage")),
	)

	ecsClient := ecs.NewFromConfig(info.Config())

	_, err := ecsClient.ListClusters(context.TODO(), &ecs.ListClustersInput{})
	require.Error(t, err)
	require.ErrorContains(t, err, "AccessDenied")

	_, err = ecsClient.ListServices(context.TODO(), &ecs.ListServicesInput{})
	require.Error(t, err)
	require.ErrorContains(t, err, "SomeCode")

}
