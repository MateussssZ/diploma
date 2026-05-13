package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	"authservice/internal/rest/handlers"
	"authservice/internal/rest/middlewares"
)

type RouterDep struct {
	Responder   handlers.IResponder
	UserHandler handlers.IUserHandlers
}

func NewRouter(dep RouterDep) http.Handler {
	r := mux.NewRouter()

	recovery := middlewares.NewRecoveryMiddleware(dep.Responder)
	r.Use(recovery.Middleware)

	auth := r.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/register", dep.UserHandler.Register).Methods(http.MethodPost)
	auth.HandleFunc("/login", dep.UserHandler.Login).Methods(http.MethodPost)
	auth.HandleFunc("/logout", dep.UserHandler.Logout).Methods(http.MethodPost)
	auth.HandleFunc("/refresh", dep.UserHandler.Refresh).Methods(http.MethodPost)

	return r
}
