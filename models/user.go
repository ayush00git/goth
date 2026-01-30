package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID			primitive.ObjectID		`json:"id" bson:"_id, omitempty"`
	UserName	string					`json:"userName" bson:"userName"`
	Email		string					`json:"email" bson:"email"`
	Role		string					`json:"role" bson:"role"`
	Password	string					`json:"password" bson:"password"`
	CreatedAt	time.Time				`json:"created_at" bson:"created_at"`
}
