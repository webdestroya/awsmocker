package awsmocker

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/smithy-go/middleware"
	"github.com/aws/smithy-go/transport/http"
)

var _ middleware.FinalizeMiddleware = (*mockerMiddleware)(nil)

func (mockerMiddleware) HandleFinalize(ctx context.Context, in middleware.FinalizeInput, next middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {

	if req, ok := in.Request.(*http.Request); ok {
		req.Header.Add(mwHeaderService, strings.ToLower(middleware.GetServiceID(ctx)))
		req.Header.Add(mwHeaderOperation, middleware.GetOperationName(ctx))

		if params := middleware.GetStackValue(ctx, mwCtxKeyParams{}); ok {
			req.Header.Add(mwHeaderParamType, fmt.Sprintf("%T", params))
		}

		if reqId, ok := middleware.GetStackValue(ctx, mwCtxKeyReqId{}).(uint64); ok {
			req.Header.Add(mwHeaderRequestId, strconv.FormatUint(reqId, 10))
		}

		in.Request = req
	}

	return next.HandleFinalize(ctx, in)
}
