package routes

import (
	"net/http"
	"goth/internals/handlers"
)

func Blog (blogHandler *handlers.BlogHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /write-blog", blogHandler.WriteBlog)
	mux.HandleFunc("GET /blogs", blogHandler.GetBlog)

	return mux
}