package routes

import (
	"net/http"
	"goth/internals/handlers"
)

func Auth (authHandler *handlers.AuthHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /signup", authHandler.Signup)
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("GET /users", authHandler.GetUsers)

	return mux
}