# AWS Mocker for Go

[![godoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/webdestroya/awsmocker)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/webdestroya/awsmocker/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/webdestroya/awsmocker)](https://goreportcard.com/report/github.com/webdestroya/awsmocker)

Easily create a proxy to allow easy testing of AWS API calls.

**:warning: This is considered alpha quality right now. It might not work for all of AWS's APIs.**

> [!IMPORTANT]  
> Version 1.0.0 has **BREAKING CHANGES**:
> * You must use the `aws.Config` struct returned in your clients.
> * The `Start` function has been modified to accept variable arguments for options setting

If you find problems, please create an Issue or make a PR.

## Installation
```shell
go get -u github.com/webdestroya/awsmocker
```

## Configuration
The default configuration when passing `nil` will setup a few mocks for STS.
```go
awsmocker.Start(t, nil)
```

For advanced usage and adding other mocks, you can use the following options:

```go
awsmocker.Start(t, &awsmocker.MockerOptions{
  // parameters
})
```

| Option Key | Type | Description |
| ----------- | ---- | ------ |
| `Mocks` | `[]*MockedEndpoint` | A list of MockedEndpoints that will be matched against all incoming requests. |
| `Timeout` | `time.Duration` | If provided, then requests that run longer than this will be terminated. Generally you should not need to set this |
| `MockEc2Metadata` | `bool` | Set this to `true` and mocks for common EC2 IMDS endpoints will be added. These are not exhaustive, so if you have a special need you will have to add it. |
| `SkipDefaultMocks` | `bool` | Setting this to true will prevent mocks for STS being added. Note: any mocks you add will be evaluated before the default mocks, so this option is generally not necessary. |
| `ReturnAwsConfig` | `bool` | For many applications, the test suite will have the ability to pass a custom `aws.Config` value. If you have the ability to do this, you can bypass setting all the HTTP_PROXY environment variables. This makes your test cleaner. Setting this to true will add the `AwsConfig` value to the returned value of the Start call. |
| `DoNotProxy` | `string` | Optional list of hostname globs that will be added to the `NO_PROXY` environment variable. These hostnames will bypass the mocker. Use this if you are making actual HTTP requests elsewhere in your code that you want to allow through. |
| `DoNotFailUnhandledRequests` | `bool` | By default, if the mocker receives any request that does not have a matching mock, it will fail the test. This is usually desired as it prevents requests without error checking from allowing tests to pass. If you explicitly want a request to fail, you can define that. |
| `DoNotOverrideCreds` | `bool` | This will stop the test mocker from overriding the AWS environment variables with fake values. This means if you do not properly configure the mocker, you could end up making real requests to AWS. This is not recommended. |

## Defining Mocks

```go
&awsmocker.MockedEndpoint{
  Request:  &awsmocker.MockedRequest{},
  Response: &awsmocker.MockedResponse{},
}
```

### Mocking Requests
| Key | Type | Description |
| --- | ---- | ---- |
| `Service`         | `string` | The AWS shortcode/subdomain for the service. (`ec2`, `ecs`, `iam`, `dynamodb`, etc) |
| `Action`          | `string` | The AWS action name that is being mocked. (`DescribeSubnets`, `ListClusters`, `AssumeRole`, etc) |
| `Params`          | `url.Values` | Matches against POST FORM PARAMs. This is only useful for older XML style API requests. This will not match against newer JSON requests. |
| `Method`          | `string` | Uppercase string of the HTTP method to match against |
| `Path`            | `string` | Matches the request path |
| `PathRegex`       | `string` | Matches the request path using regex |
| `IsEc2IMDS`       | `bool` | If set to true, then will match against the IPv4 and IPv6 hostname for EC2 IMDS |
| `JMESPathMatches` | `map[string]any` | A map of [JMESpath](https://jmespath.org/) expressions with their expected values. This will be matched against the JSON payload. |
| `Matcher`         | `func(*ReceivedRequest) bool` | A custom function that you can use to do any complex logic you want. This is run after the other matchers, so you can use them to filter down requests before they hit your matcher. |
| `MaxMatchCount`   | `int` | If this is greater than zero, then this mock will stop matching after it reaches the provided number of matches. This is useful for doing waiters. |
| `Hostname`        | `string` | Matches a specific hostname. This is normally not recommended unless you are mocking non-AWS services. |
| `Body`            | `string` | This matches a body of a request verbatim. This is not recommend unless you want to _exactly_ match a request. |
| `Strict`          | `bool` | This is only relevant if you provided `Params`. If strict mode is on, then the parameters much match entirely (and only) the provided parameter set. |

### Mocking Responses
| Key | Type | Description |
| --- | ---- | ---- |
| `Body` | SEE BELOW | SEE BELOW |
| `StatusCode` | `int` | Default 200. Allows you to override the status code for this response |
| `ContentType` | `string` | Allows overriding the content type header. By default this is handled for you based on the request |
| `Encoding` | `ResponseEncoding` | Allows you to force how the body will be encoded. By default, requests that use newer API style will receive JSON responses, and older request styles will get an XML document. |
| `DoNotWrap` | `bool` | Prevents wrapping XML responses with ACTIONResponse>ACTIONResult. Set this to true if you are doing some custom XML |
| `RootTag` | `string` | If you are doing custom XML responses, they will need a wrapping parent tag. This is where you specify the name. |
| `Handler` | `func(*ReceivedRequest) *http.Response` | If you want to handle the request entirely on your own, you can provide a function that will be passed the request and you can return an HTTP Response |

**Specifying Response Body:**

| `Body` variable type | Description |
| -------------------- | ----------- |
| `string` | Default. Body will be returned as the string verbatim |
| `func(*ReceivedRequest) (string)` | The function will be executed and the resulting string will be returned with 200OK and a default content type. |
| `func(*ReceivedRequest) (string, int)` | Same as above, but with custom status code |
| `func(*ReceivedRequest) (string, int, string)` | Same as above, but with custom content type. |
| `map` or `struct` | Will be encoded into either JSON or XML depending on the request. |


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