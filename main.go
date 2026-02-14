package main

import (
	"log"

	"goth/handlers"
	"goth/routes"
	"goth/db"

	"github.com/gin-gonic/gin"
)

func main() {

	AuthCollection := db.ConnectMongoDB("users")

	// Defining the handlers and routes
	db.CreateIndex(AuthCollection)

	authHandler := &handlers.AuthHandler{
		Collection: AuthCollection,
	}

	r := gin.Default()
	routes.AuthRoute(r, authHandler)

	log.Println("Server running on port 8080")
	log.Fatal(r.Run(":8080"))
}
