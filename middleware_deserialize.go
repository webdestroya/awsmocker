package awsmocker

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/smithy-go/middleware"
)

var _ middleware.DeserializeMiddleware = (*mockerMiddleware)(nil)

func (m *mockerMiddleware) HandleDeserialize(ctx context.Context, in middleware.DeserializeInput, next middleware.DeserializeHandler) (middleware.DeserializeOutput, middleware.Metadata, error) {
	// fmt.Printf("DESERIALIZE RAW: %s\n", in.Request)
	out, meta, err := next.HandleDeserialize(ctx, in)

	fmt.Printf("OUT.Error=%v\n", err)
	fmt.Printf("OUT.Result=%T\n", out.Result)
	fmt.Printf("OUT.RawResponse=%T\n", out.RawResponse)

	if resp, ok := out.RawResponse.(*mwResponse); ok {
		if resp.Header.Get(mwHeaderUseDB) == "true" {
			if reqId, ok := middleware.GetStackValue(ctx, mwKeyReqId{}).(uint64); ok {

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
	fmt.Printf("OUT.RawResponse=%T\n", out.RawResponse)

	return out, meta, err
}
