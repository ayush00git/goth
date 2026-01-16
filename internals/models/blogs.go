package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"

)

type Blog struct {
	ID			primitive.ObjectID		`json:"id" bson:"_id, omitempty"`
	Title		string					`json:"title" bson:"title"`
	Excerpt		string					`json:"excerpt" bson:"excerpt"`
	Tags		[]string				`json:"tags" bson:"tags"`
	Content		string					`json:"content" bson:"content"`
	IsDraft		bool					`json:"isDraft" bson:"isDraft"`
	CreatedAt	time.Time				`json:"created_at" bson:"created_at"`
}
