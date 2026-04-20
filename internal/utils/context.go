package utils

import "context"

type ctxKey string

func (c ctxKey) String() string {
	return string(c)
}

const (
	CtxRequestID     = ctxKey("XRequestID")
	CtxRequestMethod = ctxKey("method")
)

func ExtractAttrs(ctx context.Context) []any {
	var attrs []any
	if v, ok := ctx.Value(CtxRequestID).(string); ok {
		attrs = append(attrs, CtxRequestID.String(), v)
	}
	if v, ok := ctx.Value(CtxRequestMethod).(string); ok {
		attrs = append(attrs, CtxRequestMethod.String(), v)
	}

	return attrs
}
