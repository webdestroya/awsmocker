package awsmocker_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/jmespath/go-jmespath"
	"github.com/stretchr/testify/require"
	"github.com/webdestroya/awsmocker"
)

func TestEcsDescribeServices(t *testing.T) {
	awsmocker.Start(t, &awsmocker.MockerOptions{
		SkipDefaultMocks: true,
		Mocks: []*awsmocker.MockedEndpoint{
			{
				Request: &awsmocker.MockedRequest{
					Service: "ecs",
					Action:  "DescribeServices",
				},
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

	client := ecs.NewFromConfig(getAwsConfig())

	resp, err := client.DescribeServices(context.TODO(), &ecs.DescribeServicesInput{
		Services: []string{"someservice"},
		Cluster:  aws.String("testcluster"),
	})
	require.NoError(t, err)

	if *resp.Services[0].ServiceName != "someservice" {
		t.Errorf("Service name was wrong: %s", *resp.Services[0].ServiceName)
	}
}

func TestStsGetCallerIdentity_WithObj(t *testing.T) {
	awsmocker.Start(t, &awsmocker.MockerOptions{
		SkipDefaultMocks: true,
		Mocks: []*awsmocker.MockedEndpoint{
			{
				Request: &awsmocker.MockedRequest{
					Service: "sts",
					Action:  "GetCallerIdentity",
				},
				Response: &awsmocker.MockedResponse{
					Body: sts.GetCallerIdentityOutput{
						Account: aws.String(awsmocker.DefaultAccountId),
						Arn:     aws.String(fmt.Sprintf("arn:aws:iam::%s:user/fakeuser", awsmocker.DefaultAccountId)),
						UserId:  aws.String("AKIAI44QH8DHBEXAMPLE"),
					},
				},
			},
		},
	})

	stsClient := sts.NewFromConfig(getAwsConfig())

	resp, err := stsClient.GetCallerIdentity(context.TODO(), nil)
	require.NoError(t, err)

	if *resp.Account != awsmocker.DefaultAccountId {
		t.Errorf("AccountID Mismatch: %v", *resp.Account)
	}
}

func TestStsGetCallerIdentity_WithMap(t *testing.T) {
	awsmocker.Start(t, &awsmocker.MockerOptions{
		SkipDefaultMocks: true,
		Mocks: []*awsmocker.MockedEndpoint{
			{
				Request: &awsmocker.MockedRequest{
					Service: "sts",
					Action:  "GetCallerIdentity",
				},
				Response: &awsmocker.MockedResponse{
					Body: map[string]interface{}{
						"Account": awsmocker.DefaultAccountId,
						"Arn":     fmt.Sprintf("arn:aws:iam::%s:user/fakeuser", awsmocker.DefaultAccountId),
						"UserId":  "AKIAI44QH8DHBEXAMPLE",
					},
				},
			},
		},
	})
	stsClient := sts.NewFromConfig(getAwsConfig())

	resp, err := stsClient.GetCallerIdentity(context.TODO(), nil)
	require.NoError(t, err)
	require.EqualValuesf(t, awsmocker.DefaultAccountId, *resp.Account, "account id mismatch")
}

func TestDynamicMocker(t *testing.T) {
	awsmocker.Start(t, &awsmocker.MockerOptions{
		Mocks: []*awsmocker.MockedEndpoint{
			{
				Request: &awsmocker.MockedRequest{
					Service:       "events",
					Action:        "PutRule",
					MaxMatchCount: 1,
				},
				Response: &awsmocker.MockedResponse{
					Body: func(rr *awsmocker.ReceivedRequest) string {
						name, _ := jmespath.Search("Name", rr.JsonPayload)
						return awsmocker.EncodeAsJson(map[string]interface{}{
							"RuleArn": fmt.Sprintf("arn:aws:events:%s:%s:rule/%s", rr.Region, awsmocker.DefaultAccountId, name.(string)),
						})
					},
				},
			},
			{
				Request: &awsmocker.MockedRequest{
					Service:       "events",
					Action:        "PutRule",
					MaxMatchCount: 1,
				},
				Response: &awsmocker.MockedResponse{
					Body: func(rr *awsmocker.ReceivedRequest) (string, int) {
						name, _ := jmespath.Search("Name", rr.JsonPayload)
						return awsmocker.EncodeAsJson(map[string]interface{}{
							"RuleArn": fmt.Sprintf("arn:aws:events:%s:%s:rule/x%s", rr.Region, awsmocker.DefaultAccountId, name.(string)),
						}), 200
					},
				},
			},
			{
				Request: &awsmocker.MockedRequest{
					Service:       "events",
					Action:        "PutRule",
					MaxMatchCount: 1,
				},
				Response: &awsmocker.MockedResponse{
					Body: func(rr *awsmocker.ReceivedRequest) (string, int, string) {
						name, _ := jmespath.Search("Name", rr.JsonPayload)
						return awsmocker.EncodeAsJson(map[string]interface{}{
							"RuleArn": fmt.Sprintf("arn:aws:events:%s:%s:rule/y%s", rr.Region, awsmocker.DefaultAccountId, name.(string)),
						}), 200, awsmocker.ContentTypeJSON
					},
				},
			},
			{
				Request: &awsmocker.MockedRequest{
					Service:       "events",
					Action:        "PutRule",
					MaxMatchCount: 1,
				},
				Response: &awsmocker.MockedResponse{
					Body: func(rr *awsmocker.ReceivedRequest) (string, int, string, string) {
						name, _ := jmespath.Search("Name", rr.JsonPayload)
						return awsmocker.EncodeAsJson(map[string]interface{}{
							"RuleArn": fmt.Sprintf("arn:aws:events:%s:%s:rule/y%s", rr.Region, awsmocker.DefaultAccountId, name.(string)),
						}), 200, awsmocker.ContentTypeJSON, "wut"
					},
				},
			},
		},
	})

	client := eventbridge.NewFromConfig(getAwsConfig())

	tables := []struct {
		name          string
		expectedArn   string
		errorContains interface{}
	}{
		{"testrule", "arn:aws:events:us-east-1:555555555555:rule/testrule", nil},
		{"testrule", "arn:aws:events:us-east-1:555555555555:rule/xtestrule", nil},
		{"testrule", "arn:aws:events:us-east-1:555555555555:rule/ytestrule", nil},
		{"testrule", "arn:aws:events:us-east-1:555555555555:rule/ztestrule", "InvalidBodyFunc"},
	}

	for _, table := range tables {
		resp, err := client.PutRule(context.TODO(), &eventbridge.PutRuleInput{
			Name: aws.String(table.name),
		})

		if table.errorContains == nil {
			require.NoError(t, err)
			require.Equal(t, table.expectedArn, *resp.RuleArn)
		} else {
			require.ErrorContains(t, err, table.errorContains.(string))
		}
	}
}

func TestStartMockServerForTest(t *testing.T) {
	// THIS PART REALLY TALKS TO AWS
	precfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithDefaultRegion("us-east-1"),
		config.WithRetryer(func() aws.Retryer {
			return aws.NopRetryer{}
		}),
		// MAKE SURE YOU USE BAD CREDS
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("AKID", "SECRET_KEY", "TOKEN")),
	)
	require.NoError(t, err)
	_, err = sts.NewFromConfig(precfg).GetCallerIdentity(context.TODO(), nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "InvalidClientTokenId")
	// END REALLY TALKING TO AWS

	// start the test mocker server
	awsmocker.Start(t, &awsmocker.MockerOptions{})

	stsClient := sts.NewFromConfig(getAwsConfig())

	resp, err := stsClient.GetCallerIdentity(context.TODO(), nil)
	require.NoError(t, err)
	require.EqualValuesf(t, awsmocker.DefaultAccountId, *resp.Account, "account id mismatch")
}

func TestDefaultMocks(t *testing.T) {
	awsmocker.Start(t, nil)

	stsClient := sts.NewFromConfig(getAwsConfig())

	resp, err := stsClient.GetCallerIdentity(context.TODO(), nil)
	require.NoError(t, err)
	require.EqualValuesf(t, awsmocker.DefaultAccountId, *resp.Account, "account id mismatch")
}

func TestBypass(t *testing.T) {
	awsmocker.Start(t, &awsmocker.MockerOptions{
		DoNotProxy: "example.com",
	})

	httpresp, err := http.Head("http://example.com/")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpresp.StatusCode)

	stsClient := sts.NewFromConfig(getAwsConfig())

	resp, err := stsClient.GetCallerIdentity(context.TODO(), nil)
	require.NoError(t, err)
	require.EqualValuesf(t, awsmocker.DefaultAccountId, *resp.Account, "account id mismatch")

}

func TestBypassReject(t *testing.T) {
	awsmocker.Start(t, &awsmocker.MockerOptions{
		DoNotProxy: "example.com",
	})

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Head("https://example.org/")
	require.Equal(t, resp.StatusCode, http.StatusNotImplemented)
	require.ErrorContains(t, err, "Not Implemented")

	resp = nil

	resp, err = http.Get("http://example.org/")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNotImplemented, resp.StatusCode)
}

func TestSendingRegularRequestToProxy(t *testing.T) {
	info := awsmocker.Start(t, nil)

	resp, err := http.Get(info.ProxyURL + "/testing")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNotImplemented, resp.StatusCode)
}

func getAwsConfig() aws.Config {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		// add creds just in case something happens
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("XXfakekey", "XXfakesecret", "xxtoken")),
		config.WithRegion("us-east-1"),
		config.WithRetryer(func() aws.Retryer {
			return aws.NopRetryer{}
		}),
	)
	if err != nil {
		panic(err)
	}
	return cfg
}
