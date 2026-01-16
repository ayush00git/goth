package main

import (
	"context"
	"net/http"
	"log"
	"fmt"
	"os"
	"time"

	"goth/internals/handlers"
	"goth/internals/routes"

	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(`Unable to locate the .env`)
	}

	uri := os.Getenv("MONGO_URI");
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

	collection := client.Database("goth_db").Collection("users")

	authHandler := &handlers.AuthHandler{
		Collection: collection,
	}

	srv := &http.Server{
		Addr: ":8080",
		Handler: routes.Auth(authHandler),
	}
	
	fmt.Println("Server started on port 8080")
	log.Fatal(srv.ListenAndServe());
}