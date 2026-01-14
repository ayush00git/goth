package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID			primitive.ObjectID		`json:"id" bson:"id, omitempty"`
	Email		string					`json:"email" bson:"email"`
	Password	string					`json:"password" bson:"password"`
	CreatedAt	time.Time				`json:"created_at" bson:"created_at"`
};
