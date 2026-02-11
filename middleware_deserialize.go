package awsmocker

import (
	"context"
	"errors"

	"github.com/aws/smithy-go/middleware"
)

var _ middleware.DeserializeMiddleware = (*mockerMiddleware)(nil)

func (m *mockerMiddleware) HandleDeserialize(ctx context.Context, in middleware.DeserializeInput, next middleware.DeserializeHandler) (middleware.DeserializeOutput, middleware.Metadata, error) {
	out, meta, err := next.HandleDeserialize(ctx, in)

	if resp, ok := out.RawResponse.(*mwResponse); ok {
		if resp.Header.Get(mwHeaderUseDB) == "true" {
			if reqId, ok := middleware.GetStackValue(ctx, mwCtxKeyReqId{}).(uint64); ok {

				res, _ := m.mocker.requestLog.LoadAndDelete(reqId)
				entry := res.(mwDBEntry)

				return middleware.DeserializeOutput{
					RawResponse: resp,
					Result:      entry.Response,
				}, meta, entry.Error

			} else {
				return out, meta, errors.New("invalid mocker result?")
			}
		}
	}

	return out, meta, err
}
