package awsmocker

import (
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

const (
	mwHeaderService   = `X-Awsmocker-Service`
	mwHeaderOperation = `X-Awsmocker-Operation`
	mwHeaderParamType = `X-Awsmocker-Param-Type`
	mwHeaderRequestId = `X-Awsmocker-Request-Id`
	mwHeaderError     = `X-Awsmocker-Error`
	mwHeaderUseDB     = `X-Awsmocker-Use-Db`
)

type (
	mwCtxKeyReqId  struct{}
	mwCtxKeyParams struct{}
	mwCtxKeyUseDB  struct{}
)

type (
	mwRequest  = smithyhttp.Request
	mwResponse = smithyhttp.Response
)

type mockerMiddleware struct {
	mocker *mocker
}

func (mockerMiddleware) ID() string {
	return "awsmocker"
}

func addMiddlewareConfigOption(m *mocker) AwsLoadOptionsFunc {

	mockerMW := &mockerMiddleware{
		mocker: m,
	}

	return config.WithAPIOptions([]func(*middleware.Stack) error{
		func(stack *middleware.Stack) error {
			if err := stack.Initialize.Add(mockerMW, middleware.After); err != nil {
				return err
			}

			if err := stack.Serialize.Add(mockerMW, middleware.After); err != nil {
				return err
			}

			if err := stack.Build.Add(mockerMW, middleware.After); err != nil {
				return err
			}

			if err := stack.Deserialize.Add(mockerMW, middleware.Before); err != nil {
				return err
			}

			if err := stack.Finalize.Add(mockerMW, middleware.After); err != nil {
				return err
			}
			return nil
		},
	})
}
