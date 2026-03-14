package routes

import (
	"github.com/ayush00git/goth/handlers"
	"github.com/ayush00git/goth/middlewares"

	"github.com/gin-gonic/gin"
)

func AuthRoute(router *gin.Engine, authHandler *handlers.AuthHandler) {
	api := router.Group("/api/auth")
	{
		api.POST("/signup", authHandler.Signup)
		api.GET("/verify", authHandler.VerifyEmail)
		api.POST("/login", authHandler.Login)
		api.POST("/logout", middlewares.AuthMiddleware(), authHandler.Logout)
		api.GET("/users", middlewares.AuthMiddleware(), authHandler.GetUsers)
	}
}
