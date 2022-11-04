package awsmocker_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/webdestroya/awsmocker"
)

func init() {
	awsmocker.GlobalDebugMode = true
}

// just make sure our link hack (resetProxyConfig) works
func TestResetProxyEnvHack(t *testing.T) {
	closeMocker, _, _ := awsmocker.StartMockServer(&awsmocker.MockerOptions{
		T: t,
	})
	closeMocker()
}

func TestEcsDescribeServices(t *testing.T) {
	closeMocker, _, _ := awsmocker.StartMockServer(&awsmocker.MockerOptions{
		T:                t,
		Verbose:          true,
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
	defer closeMocker()

	client := ecs.NewFromConfig(getAwsConfig())

	resp, err := client.DescribeServices(context.TODO(), &ecs.DescribeServicesInput{
		Services: []string{"someservice"},
		Cluster:  aws.String("testcluster"),
	})
	if err != nil {
		t.Errorf("Error ECS.DescribeServices: %s", err)
		return
	}

	if *resp.Services[0].ServiceName != "someservice" {
		t.Errorf("Service name was wrong: %s", *resp.Services[0].ServiceName)
	}

	_ = resp
}

func TestStsGetCallerIdentity_WithObj(t *testing.T) {
	closeMocker, _, _ := awsmocker.StartMockServer(&awsmocker.MockerOptions{
		T:                t,
		Verbose:          true,
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
	defer closeMocker()

	stsClient := sts.NewFromConfig(getAwsConfig())

	resp, err := stsClient.GetCallerIdentity(context.TODO(), nil)
	if err != nil {
		t.Errorf("Error STS.GetCallerIdentity: %s", err)
		return
	}

	if *resp.Account != awsmocker.DefaultAccountId {
		t.Errorf("AccountID Mismatch: %v", *resp.Account)
	}
}

func TestStsGetCallerIdentity_WithMap(t *testing.T) {
	closeMocker, _, _ := awsmocker.StartMockServer(&awsmocker.MockerOptions{
		T:                t,
		Verbose:          true,
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
	defer closeMocker()

	stsClient := sts.NewFromConfig(getAwsConfig())

	resp, err := stsClient.GetCallerIdentity(context.TODO(), nil)
	if err != nil {
		t.Errorf("Error STS.GetCallerIdentity: %s", err)
		return
	}

	if *resp.Account != awsmocker.DefaultAccountId {
		t.Errorf("AccountID Mismatch: %v", *resp.Account)
	}
}

func TestDefaultMocks(t *testing.T) {
	closeMocker, _, _ := awsmocker.StartMockServer(&awsmocker.MockerOptions{
		T:       t,
		Verbose: true,
	})
	defer closeMocker()

	stsClient := sts.NewFromConfig(getAwsConfig())

	resp, err := stsClient.GetCallerIdentity(context.TODO(), nil)
	if err != nil {
		t.Errorf("Error STS.GetCallerIdentity: %s", err)
		return
	}

	if *resp.Account != awsmocker.DefaultAccountId {
		t.Errorf("AccountID Mismatch: %v", *resp.Account)
	}
}

func getAwsConfig() aws.Config {
	cfg, err := config.LoadDefaultConfig(context.TODO(), func(lo *config.LoadOptions) error {
		lo.Region = "us-east-1"
		return nil
	})
	if err != nil {
		panic(err)
	}
	return cfg
}
