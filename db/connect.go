package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ayush00git/goth/helpers"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongoDB(collection string) *mongo.Collection {
	uri := helpers.GetEnvVar("MONGO_URI")

	// context, client, ping
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Connection to MongoDB established")

	collec := client.Database("goth").Collection(collection)
	return collec
}
