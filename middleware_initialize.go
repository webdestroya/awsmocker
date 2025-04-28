package awsmocker

import (
	"context"

	"github.com/aws/smithy-go/middleware"
)

var _ middleware.InitializeMiddleware = (*mockerMiddleware)(nil)

func (m *mockerMiddleware) HandleInitialize(ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler) (middleware.InitializeOutput, middleware.Metadata, error) {

	// if _, ok := in.Parameters.(*ec2.DescribeSubnetsInput); ok {
	// 	out := middleware.InitializeOutput{
	// 		Result: &ec2.DescribeSubnetsOutput{
	// 			Subnets: []ec2Types.Subnet{
	// 				{
	// 					SubnetId: aws.String("subnet-aaaaaaaa"),
	// 					VpcId:    aws.String("vpc-11111111"),
	// 				},
	// 			},
	// 		},
	// 	}
	// 	return out, middleware.Metadata{}, nil
	// }

	reqId := m.mocker.mwReqCounter.Add(1)

	ctx = withMWContextParam(ctx, in.Parameters)
	ctx = middleware.WithStackValue(ctx, mwKeyReqId{}, reqId)

	return next.HandleInitialize(ctx, in)
}
