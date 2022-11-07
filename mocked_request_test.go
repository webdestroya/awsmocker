package awsmocker

import (
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMockedRequest_Inspect(t *testing.T) {
	mr := &MockedRequest{
		Strict:    true,
		Hostname:  "awsmocker.local",
		Service:   "sts",
		Action:    "GetCallerIdentity",
		Body:      "somebody",
		Method:    http.MethodGet,
		Path:      "/blah/blah/some/path",
		PathRegex: regexp.MustCompile(`/version/([a-z0-9]+)/test`),
		IsEc2IMDS: true,
		Params: url.Values{
			"TestParam": []string{"thing"},
		},
	}

	result := mr.Inspect()

	require.Contains(t, result, "STRICT")
	require.Contains(t, result, "Service=sts")
	require.Contains(t, result, "Action=GetCallerIdentity")
	require.Contains(t, result, "imds=true")
	require.Contains(t, result, "Hostname=awsmocker.local")
	require.Contains(t, result, "Path=/blah/blah/some/path")
	require.Contains(t, result, "Method=GET")
	require.Contains(t, result, "Params=TestParam=thing")
	require.Contains(t, result, "Body=somebody")
}
