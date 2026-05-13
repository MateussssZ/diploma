package middlewares

import (
	"fmt"
	"io"
	"net/http"

	"authservice/internal/rest/handlers"
)

type RecoveryMiddleware struct {
	responder handlers.IResponder
}

func NewRecoveryMiddleware(responder handlers.IResponder) *RecoveryMiddleware {
	return &RecoveryMiddleware{responder: responder}
}

func (mw *RecoveryMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					body = []byte("cannot read body")
				}
				mw.responder.WriteError(r.Context(), w,
					fmt.Errorf("http handler panicked"),
					handlers.WithStatusCode(http.StatusInternalServerError),
					handlers.WithRequestBody(string(body)),
					handlers.WithTags("method", r.Method),
					handlers.WithTags("url", r.URL.String()),
					handlers.WithTags("panic", rec),
				)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
