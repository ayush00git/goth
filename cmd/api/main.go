package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"goth/internals/handlers"
	"goth/internals/routes"
	"goth/internals/db"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
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

	AuthCollection := client.Database("goth_db").Collection("users")
	BlogCollection := client.Database("goth_db").Collection("blogs")

	db.CreateIndex(AuthCollection)

	authHandler := &handlers.AuthHandler{
		Collection: AuthCollection,
	}

	blogHandler := &handlers.BlogHandler{
		Collection: BlogCollection,
	}

	mux := http.NewServeMux()
	mux.Handle("/auth/", routes.Auth(authHandler))
	mux.Handle("/blog/", routes.Blog(blogHandler))

	srv := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}

	fmt.Println("Server started on port 8080")
	log.Fatal(srv.ListenAndServe())
}
