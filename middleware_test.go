package awsmocker

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/require"
)

func TestSerializeMiddleware(t *testing.T) {

	m := Start(t,
		WithMocks(&MockedEndpoint{
			Request: &MockedRequest{
				Service: "ecs",
				Action:  "DescribeServices",
			},
			Response: &MockedResponse{
				Body: map[string]any{
					"services": []map[string]any{
						{
							"serviceName": "someservice",
						},
					},
				},
			},
		}))

	client := ecs.NewFromConfig(m.Config())
	_, _ = sts.NewFromConfig(m.Config()).GetCallerIdentity(context.TODO(), nil)

	resp, err := client.DescribeServices(context.TODO(), &ecs.DescribeServicesInput{
		Services: []string{"someservice"},
		Cluster:  aws.String("testcluster"),
	})
	require.NoError(t, err)
	require.Equalf(t, "someservice", *resp.Services[0].ServiceName, "Service name was wrong")
}

func TestEC2Middleware(t *testing.T) {

	m := Start(t,
		WithMocks(&MockedEndpoint{
			Request: &MockedRequest{
				Service: "ec2",
				Action:  "DescribeSubnets",
			},
			// Response: &MockedResponse{
			// 	DoNotWrap: true,
			// 	Body: map[string]any{
			// 		"requestId": "43e9cb52-0e10-40fe-b457-988c8fbfea26",
			// 		"subnetSet": map[string]any{
			// 			"item": []any{
			// 				map[string]any{
			// 					"subnetId": "subnet-633333333333",
			// 					"vpcId":    "vpc-123456789",
			// 				},
			// 				map[string]any{
			// 					"subnetId": "subnet-644444444444",
			// 					"vpcId":    "vpc-123456789",
			// 				},
			// 			},
			// 		},
			// 	},
			// },
			Response: &MockedResponse{
				Body: &ec2.DescribeSubnetsOutput{
					Subnets: []ec2Types.Subnet{
						{
							VpcId: aws.String("vpc-123456789"),
						},
					},
				},
			},
		}))

	client := ec2.NewFromConfig(m.Config())

	resp, err := client.DescribeSubnets(context.TODO(), &ec2.DescribeSubnetsInput{})
	require.NoError(t, err)
	require.NotNil(t, resp)

	require.GreaterOrEqual(t, len(resp.Subnets), 1)
	require.Equal(t, "vpc-123456789", *resp.Subnets[0].VpcId)
}
