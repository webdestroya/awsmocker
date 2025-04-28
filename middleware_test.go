package awsmocker

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go/document"
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
				Body: "BLAH",
			},
		}))

	client := ec2.NewFromConfig(m.Config())

	resp, err := client.DescribeSubnets(context.TODO(), &ec2.DescribeSubnetsInput{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.GreaterOrEqual(t, len(resp.Subnets), 1)
	require.Equal(t, "vpc-123456789", *resp.Subnets[0].VpcId)
}

func TestReflection(t *testing.T) {
	thing := &ec2.DescribeAccountAttributesOutput{
		AccountAttributes: []ec2Types.AccountAttribute{
			{
				AttributeName: aws.String("blah"),
				AttributeValues: []ec2Types.AccountAttributeValue{
					{
						AttributeValue: aws.String("yar"),
					},
				},
			},
		},
	}

	f1 := func(rr *ReceivedRequest) (*ec2.DescribeSubnetsOutput, error) {
		return nil, nil
	}

	typ := reflect.TypeOf(thing)

	require.True(t, document.IsNoSerde(thing))

	t.Logf("TYPE=%s", typ.String())
	t.Logf("TYPE=%s", typ.Elem().String())
	t.Logf("PKG=%s", typ.Elem().PkgPath())

	f1type := reflect.TypeOf(f1)
	t.Logf("F1 TYPE=%s", f1type.String())
	t.Logf("F1 NumIn=%d", f1type.NumIn())
	t.Logf("F1 NumOut=%d", f1type.NumOut())
	t.Logf("F1 In.0=%s", f1type.In(0).String())
	t.Logf("F1 Out.0=%s [%s]", f1type.Out(0).String(), f1type.Out(0).Elem().PkgPath())

}
