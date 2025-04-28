package awsmocker

import (
	"context"

	"github.com/aws/smithy-go/middleware"
)

var _ middleware.BuildMiddleware = (*mockerMiddleware)(nil)

func (mockerMiddleware) HandleBuild(ctx context.Context, in middleware.BuildInput, next middleware.BuildHandler) (middleware.BuildOutput, middleware.Metadata, error) {
	return next.HandleBuild(ctx, in)
}
