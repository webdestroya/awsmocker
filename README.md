# AWS Mocker

Easily create a proxy to allow easy testing of AWS API calls.

**Warning! This is considered alpha quality right now. It might not work for all of AWS's APIs.**

If you find problems, please create an Issue or make a PR.



## Usage

```go
func TestSomethingThatCallsAws(t *testing.T) {
	closeMocker, _, _ := awsmocker.StartMockServer(&awsmocker.MockerOptions{
		T: t,

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
					Body: ecs.DescribeServicesOutput{
						Services: []ecstypes.Service{
							{
								ServiceName: aws.String("someservice"),
							},
						},
					},
				},
			},
		},
	})
	defer closeMocker()

	cfg, _ := config.LoadDefaultConfig(context.TODO(), func(lo *config.LoadOptions) error {
		lo.Region = "us-east-1"
		return nil
	})

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

## Assumptions/Limitations
* The first matching mock is returned.
* Service is assumed by the credential header
* Action is calculated by the `Action` parameter, or the `X-amz-target` header.
* if you provide a response object, it will be encoded to JSON or XML based on the requesting content type. If you need a response in a special format, please provide the content type and a string for the body.
* There is very little "error handling". If something goes wrong, it just panics. This might be less than ideal, but the only usecase for this library is within a test, which would make the test fail. This is the goal.

## See Also
* Heavily influenced by [hashicorp's servicemocks](github.com/hashicorp/aws-sdk-go-base/v2/servicemocks)