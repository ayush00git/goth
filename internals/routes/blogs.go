package routes

import (
	"goth/internals/handlers"
	"net/http"
)

func Blog(blogHandler *handlers.BlogHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /blog/write", blogHandler.WriteBlog)
	mux.HandleFunc("GET /blog/", blogHandler.GetBlog)
	mux.HandleFunc("GET /blog/{BlogID}", blogHandler.GetBlogByID)
	mux.HandleFunc("DELETE /blog/{BlogID}", blogHandler.DeleteBlogByID)

	return mux
}
