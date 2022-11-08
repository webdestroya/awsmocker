# AWS Mocker for Go

[![godoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/webdestroya/awsmocker)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/webdestroya/awsmocker/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/webdestroya/awsmocker)](https://goreportcard.com/report/github.com/webdestroya/awsmocker)

Easily create a proxy to allow easy testing of AWS API calls.

**:warning: This is considered alpha quality right now. It might not work for all of AWS's APIs.**

If you find problems, please create an Issue or make a PR.

## Installation
```shell
go get -u github.com/webdestroya/awsmocker
```

## Usage

```go
func TestSomethingThatCallsAws(t *testing.T) {
	awsmocker.Start(t, &awsmocker.MockerOptions{
		// List out the mocks
		Mocks: []*awsmocker.MockedEndpoint{
			// Simple construction of a response
			awsmocker.NewSimpleMockedEndpoint("sts", "GetCallerIdentity", sts.GetCallerIdentityOutput{
				Account: aws.String("123456789012"),
				Arn:     aws.String("arn:aws:iam::123456789012:user/fakeuser"),
				UserId:  aws.String("AKIAI44QH8DHBEXAMPLE"),
			}),

			// advanced construction
			{
				Request: &awsmocker.MockedRequest{
					// specify the service/action to respond to
					Service: "ecs",
					Action:  "DescribeServices",
				},
				// provide the response to give
				Response: &awsmocker.MockedResponse{
					Body: map[string]interface{}{
						"services": []map[string]interface{}{
							{
								"serviceName": "someservice",
							},
						},
					},
				},
			},
		},
	})

	cfg, _ := config.LoadDefaultConfig(context.TODO())

	stsClient := sts.NewFromConfig(cfg)

	resp, err := stsClient.GetCallerIdentity(context.TODO(), nil)
	if err != nil {
		t.Errorf("Error STS.GetCallerIdentity: %s", err)
		return
	}

	if *resp.Account != "123456789012" {
		t.Errorf("AccountID Mismatch: %v", *resp.Account)
	}

	// ... do the rest of your test here
}
```

## Defining Mocks

### Dynamic Response
```go
func Mock_Events_PutRule_Generic() *awsmocker.MockedEndpoint {
	return &awsmocker.MockedEndpoint{
		Request: &awsmocker.MockedRequest{
			Service: "events",
			Action:  "PutRule",
		},
		Response: &awsmocker.MockedResponse{
			Body: func(rr *awsmocker.ReceivedRequest) string {

				name, _ := jmespath.Search("Name", rr.JsonPayload)

				return util.Must(util.Jsonify(map[string]interface{}{
					"RuleArn": fmt.Sprintf("arn:aws:events:%s:%s:rule/%s", rr.Region, awsmocker.DefaultAccountId, name.(string)),
				}))
			},
		},
	}
}
```

## Viewing Requests/Responses

To see the request/response traffic, you can use either of the following:

* Set `awsmocker.GlobalDebugMode = true` in your tests
* Use the `AWSMOCKER_DEBUG=true` environment variable

## Assumptions/Limitations
* The first matching mock is returned.
* Service is assumed by the credential header
* Action is calculated by the `Action` parameter, or the `X-amz-target` header.
* if you provide a response object, it will be encoded to JSON or XML based on the requesting content type. If you need a response in a special format, please provide the content type and a string for the body.
* There is very little "error handling". If something goes wrong, it just panics. This might be less than ideal, but the only usecase for this library is within a test, which would make the test fail. This is the goal.

## See Also
* Heavily influenced by [hashicorp's servicemocks](github.com/hashicorp/aws-sdk-go-base/v2/servicemocks)