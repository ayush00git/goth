package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"goth/handlers"
	"goth/routes"
	"goth/db"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {

	// getting MONGO_URI from env and connecting
	// to the database
	if err := godotenv.Load(); err != nil {
		log.Fatal(`Unable to locate the .env`)
	}

	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		log.Fatal("MONGO_URI string is not present in the env variables")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Connection to MongoDB is failing")
	}
	fmt.Println("MongoDB connected successfully!")

	// Defining the handlers and routes
	AuthCollection := client.Database("goth_db").Collection("users")
	db.CreateIndex(AuthCollection)

	authHandler := &handlers.AuthHandler{
		Collection: AuthCollection,
	}

	r := gin.Default()
	routes.AuthRoute(r, authHandler)

	log.Println("Server running on port 8080")
	log.Fatal(r.Run(":8080"))
}
