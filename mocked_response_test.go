package awsmocker

import (
	"maps"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/require"
)

func TestMockedResponse_getResponse(t *testing.T) {

	fakePayload := map[string]any{
		"foo": map[string]any{
			"baz": "thing",
		},
		"bar": 1234,
		"b":   true,
		"bf":  false,
		"str": "some string",
	}

	tables := []struct {
		name string
		mr   *MockedResponse
		rr   *ReceivedRequest
		exp  func(*testing.T, *httpResponse)
	}{
		{
			name: "XMLWrapped",
			mr: &MockedResponse{
				Body:     maps.Clone(fakePayload),
				Encoding: ResponseEncodingXML,
			},
			rr: &ReceivedRequest{
				Service: "sts",
				Action:  "GetCallerIdentity",
			},
			exp: func(t *testing.T, hr *httpResponse) {
				require.Equal(t, ContentTypeXML, hr.contentType)
				bodyRaw := string(hr.bodyRaw)
				require.Contains(t, bodyRaw, "<GetCallerIdentityResult>")
				require.Contains(t, bodyRaw, "<GetCallerIdentityResponse>")
			},
		},
	}

	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			hr := table.mr.getResponse(table.rr)
			table.exp(t, hr)
		})
	}
}

func TestProcessDirectRequestFunc(t *testing.T) {

	entry := mwDBEntry{
		Parameters: &sts.GetCallerIdentityInput{},
	}
	rr := &ReceivedRequest{}

	stsResponse := &sts.GetCallerIdentityOutput{
		Account: aws.String(DefaultAccountId),
		Arn:     aws.String("arn"),
		UserId:  aws.String("userid"),
	}

	tables := []struct {
		fn any
	}{
		{
			fn: func(_ *sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error) {
				return stsResponse, nil
			},
		},

		{
			fn: func(_ *ReceivedRequest) (*sts.GetCallerIdentityOutput, error) {
				return stsResponse, nil
			},
		},

		{
			fn: func(_ *ReceivedRequest, _ *sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error) {
				return stsResponse, nil
			},
		},

		{
			fn: func() *sts.GetCallerIdentityOutput {
				return stsResponse
			},
		},
	}

	for _, table := range tables {
		t.Run("entry", func(t *testing.T) {
			result, err := processDirectRequestFunc(entry, rr, reflect.ValueOf(table.fn))
			_ = result
			_ = err
		})
	}
}
