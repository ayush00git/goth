package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID			primitive.ObjectID		`json:"id" bson:"_id, omitempty"`
	UserName	string					`json:"userName" bson:"userName" binding:"required"`
	Email		string					`json:"email" bson:"email" binding:"required"`
	Role		string					`json:"role" bson:"role"`
	Password	string					`json:"password" bson:"password" binding:"required"`
	CreatedAt	time.Time				`json:"created_at" bson:"created_at"`
}
