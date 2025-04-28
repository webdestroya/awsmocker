package awsmocker

import (
	"context"

	"github.com/aws/smithy-go/middleware"
)

var _ middleware.SerializeMiddleware = (*mockerMiddleware)(nil)

func (mockerMiddleware) HandleSerialize(ctx context.Context, in middleware.SerializeInput, next middleware.SerializeHandler) (middleware.SerializeOutput, middleware.Metadata, error) {
	return next.HandleSerialize(ctx, in)
}
