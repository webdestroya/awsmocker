# AWS Mocker for Go

[![godoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/webdestroya/awsmocker)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/webdestroya/awsmocker/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/webdestroya/awsmocker)](https://goreportcard.com/report/github.com/webdestroya/awsmocker)

Easily create a proxy to allow easy testing of AWS API calls.

**:warning: This is considered alpha quality right now. It might not work for all of AWS's APIs.**

> [!IMPORTANT]  
> Version 1.0.0 has **BREAKING CHANGES**:
> * You must use the AWS config from `Config()` function returned by `Start()`.
> * The `Start` function has been modified to accept variable arguments for options setting

If you find problems, please create an Issue or make a PR.

## Installation
```shell
go get -u github.com/webdestroya/awsmocker
```

## Configuration
The default configuration will setup a few mocks for STS.
```go
m := awsmocker.Start(t)
```

For advanced usage and adding other mocks, you can use the following options:

```go
m := awsmocker.Start(t, ...OPTIONS)
```

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
  m := awsmocker.Start(t, awsmocker.WithMocks(
      // Simple construction of a response
      awsmocker.NewSimpleMockedEndpoint("sts", "GetCallerIdentity", sts.GetCallerIdentityOutput{
        Account: aws.String("123456789012"),
        Arn:     aws.String("arn:aws:iam::123456789012:user/fakeuser"),
        UserId:  aws.String("AKIAI44QH8DHBEXAMPLE"),
      }),

      // advanced construction
      &awsmocker.MockedEndpoint{
        Request: &awsmocker.MockedRequest{
          // specify the service/action to respond to
          Service: "ecs",
          Action:  "DescribeServices",
        },
        // provide the response to give
        Response: &awsmocker.MockedResponse{
          Body: map[string]any{
            "services": []map[string]any{
              {
                "serviceName": "someservice",
              },
            },
          },
        },
      },
    ),
  )

  stsClient := sts.NewFromConfig(m.Config())

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

        return util.Must(util.Jsonify(map[string]any{
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

## Possible Issues

**Receiving error: "not found, ResolveEndpointV2"**:  
Upgrade aws modules: `go get -u github.com/aws/aws-sdk-go-v2/...`

## See Also
* Heavily influenced by [hashicorp's servicemocks](github.com/hashicorp/aws-sdk-go-base/v2/servicemocks)