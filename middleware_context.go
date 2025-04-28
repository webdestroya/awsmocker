package awsmocker

import "context"

type ctxKeyT string

const (
	ctxKeyParam = ctxKeyT("param")
)

func withMWContextParam(ctx context.Context, param any) context.Context {
	return context.WithValue(ctx, ctxKeyParam, param)
}

func getMWContextParam(ctx context.Context) any {
	return ctx.Value(ctxKeyParam)
}
