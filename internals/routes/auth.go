package routes

import (
	"net/http"
	"goth/internals/handlers"
)

func Auth (authHandler *handlers.AuthHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /auth/signup", authHandler.Signup)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("GET /auth/users", authHandler.GetUsers)

	return mux
}