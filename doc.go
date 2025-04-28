// Package awsmocker allows easier mocking of AWS API responses.
//
// # Example Usage
//
// The following is a complete example using awsmocker in an example test:
//
//	import (
//	  "testing"
//		"context"
//
//		"github.com/aws/aws-sdk-go-v2/aws"
//		"github.com/aws/aws-sdk-go-v2/config"
//		"github.com/aws/aws-sdk-go-v2/service/ecs"
//	  "github.com/webdestroya/awsmocker"
//	)
//
//
//	func TestEcsDescribeServices(t *testing.T) {
//		m := awsmocker.Start(t, awsmocker.WithMocks(&awsmocker.MockedEndpoint{
//			Request: &awsmocker.MockedRequest{
//				Service: "ecs",
//				Action:  "DescribeServices",
//			},
//			Response: &awsmocker.MockedResponse{
//				Body: map[string]any{
//					"services": []map[string]any{
//						{
//							"serviceName": "someservice",
//						},
//					},
//				},
//			},
//		}))
//
//		client := ecs.NewFromConfig(m.Config())
//
//			resp, err := client.DescribeServices(context.TODO(), &ecs.DescribeServicesInput{
//				Services: []string{"someservice"},
//				Cluster:  aws.String("testcluster"),
//			})
//			if err != nil {
//				t.Errorf(err)
//			}
//			if *resp.Services[0].ServiceName != "someservice" {
//				t.Errorf("Service name was wrong")
//			}
//		}
package awsmocker
