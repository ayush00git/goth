package routes

import (
	"net/http"
	"goth/internals/handlers"
)

func Auth(authHandler *handlers.AuthHandler) http.Handler {
	mux := http.NewServeMux();

	mux.HandleFunc("POST /signup", authHandler.SignUp);
	mux.HandleFunc("POST /login", authHandler.LogIn);

	return mux;
}