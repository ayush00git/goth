package routes

import (
	"goth/handlers"

	"github.com/gin-gonic/gin"
)

func AuthRoute (router *gin.Engine, authHandler *handlers.AuthHandler) {
	api := router.Group("/api/auth")
	{
		api.POST("/signup", authHandler.Signup)
		api.POST("/login", authHandler.Login)
		api.GET("/users", authHandler.GetUsers)
		api.POST("/logout", authHandler.Logout)
	}
}
