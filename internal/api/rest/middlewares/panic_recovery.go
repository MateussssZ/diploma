package middlewares

import (
	"fmt"
	"apigateway/internal/api/rest/handlers"
	"apigateway/internal/pkg/errorspkg"
	"apigateway/internal/pkg/validate"
	"io"
	"net/http"
)

type RecoveryMiddlewareDep struct {
	Responder handlers.IResponder `validate:"required"`
}

type RecoveryMiddleware struct {
	responder handlers.IResponder
}

func NewRecoveryMiddleware(dep RecoveryMiddlewareDep) (*RecoveryMiddleware, error) {
	if err := validate.Struct(dep); err != nil {
		return nil, errorspkg.NewValidationError("NewRecoveryMiddleware", err)
	}

	return &RecoveryMiddleware{
		responder: dep.Responder,
	}, nil
}

func (mw *RecoveryMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if rec := recover(); rec != nil {
				jsonBody, err := io.ReadAll(r.Body)
				if err != nil {
					jsonBody = []byte("cannot marshal body")
				}

				mw.responder.WriteError(r.Context(), w, fmt.Errorf("http handler panicked"),
					handlers.WithStatusCode(http.StatusInternalServerError),
					handlers.WithRequestBody(string(jsonBody)),
					handlers.WithTags("method", r.Method),
					handlers.WithTags("url", r.URL.String()),
					handlers.WithTags("panic", rec),
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
