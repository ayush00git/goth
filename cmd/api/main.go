package main

import (
	"context"
	"net/http"
	"fmt"
	"log"
	"os"
	"time"

	"goth/internals/handlers"
	"goth/internals/routes"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env exists for this project");
	}

	uri := os.Getenv("MONGO_URI");
	if uri == "" {
		log.Fatal("MONGO_URI string is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second);
	defer cancel();

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri));
	if err != nil {
		log.Fatal(err);
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Connection to MongoDB failed");
	}
	fmt.Println("Connection to MongoDB successfull!");

	collection := client.Database("goth_db").Collection("users");

	authHandler := &handlers.AuthHandler {
		Collection: collection,
	}

	srv := &http.Server{
		Addr: ":8080",
		Handler: routes.Auth(authHandler),
	}

	fmt.Println("Server started on port 8080");
	log.Fatal(srv.ListenAndServe());
}