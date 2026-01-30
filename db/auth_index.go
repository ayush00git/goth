package db

import (
	"context"
	"log"
	"time"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateIndex(collection *mongo.Collection) {
	emailIndex := mongo.IndexModel {
		Keys: bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	userNameIndex := mongo.IndexModel {
		Keys: bson.D{{Key: "userName", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.Indexes().CreateOne(ctx, emailIndex)
	if err != nil {
		log.Fatal("Error setting email as index", err)
	}

	_, err = collection.Indexes().CreateOne(ctx, userNameIndex)
	if err != nil {
		log.Fatal("Error setting userName as index", err)
	}

	fmt.Println("Created email and userName index")
}
