package awsmocker

import (
	"fmt"
	"net/http"
)

var (
	MockStsGetCallerIdentityValid = &MockedEndpoint{
		Request: &MockedRequest{
			Service: "sts",
			Action:  "GetCallerIdentity",
		},
		Response: &MockedResponse{
			StatusCode: http.StatusOK,
			Encoding:   ResponseEncodingXML,
			Body: map[string]interface{}{
				"Account": DefaultAccountId,
				"Arn":     fmt.Sprintf("arn:aws:iam::%s:user/fakeuser", DefaultAccountId),
				"UserId":  "AKIAI44QH8DHBEXAMPLE",
			},
		},
	}
)
