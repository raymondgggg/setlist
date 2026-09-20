package graph

import (
	"context"
	"net/http"
)

type contextKey string

const responseWriterKey contextKey = "responseWriter"
const requestReaderKey contextKey = "requestReader"

func WithResponseWriter(ctx context.Context, w http.ResponseWriter) context.Context {
	return context.WithValue(ctx, responseWriterKey, w)
}

func ResponseWriter(ctx context.Context) (http.ResponseWriter, bool) {
	w, ok := ctx.Value(responseWriterKey).(http.ResponseWriter)
	return w, ok
}

func WithRequestReader(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, requestReaderKey, r)
}

func RequestReader(ctx context.Context) (*http.Request, bool) {
	r, ok := ctx.Value(requestReaderKey).(*http.Request)
	return r, ok
}
