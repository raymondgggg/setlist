package graph

import (
	"context"
	"net/http"
)

type contextKey string

const responseWriterKey contextKey = "responseWriter"

func WithResponseWriter(ctx context.Context, w http.ResponseWriter) context.Context {
	return context.WithValue(ctx, responseWriterKey, w)
}

func ResponseWriter(ctx context.Context) (http.ResponseWriter, bool) {
	w, ok := ctx.Value(responseWriterKey).(http.ResponseWriter)
	return w, ok
}
