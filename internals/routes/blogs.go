package routes

import (
	"goth/internals/handlers"
	"goth/internals/middlewares"
	"net/http"
)

func Blog(blogHandler *handlers.BlogHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /blog/", blogHandler.GetBlog)

	mux.Handle("POST /blog/write/", middlewares.AuthMiddleware(http.HandlerFunc(blogHandler.WriteBlog)))
	mux.Handle("DELETE /blog/delete/{BlogID}", middlewares.AuthMiddleware(http.HandlerFunc(blogHandler.DeleteBlogByID)))
	mux.Handle("PUT /blog/edit/{BlogID}", middlewares.AuthMiddleware(http.HandlerFunc(blogHandler.EditBlogByID)))

	mux.Handle("GET /blog/{BlogID}/", middlewares.AuthMiddleware(http.HandlerFunc(blogHandler.GetBlogByID)))

	return mux
}
