package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"goth/internals/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type BlogHandler struct {
	Collection *mongo.Collection
}

func (h *BlogHandler) WriteBlog (w http.ResponseWriter, r *http.Request) {
	var blog models.Blog

	blog.ID = primitive.NewObjectID()
	blog.CreatedAt = time.Now()

	if err := json.NewDecoder(r.Body).Decode(&blog); err != nil {
		http.Error(w, "Error fetching user's request", http.StatusBadRequest)
		return
	}

	_, err := h.Collection.InsertOne(context.TODO(), blog)
	if err != nil {
		http.Error(w, "Error submitting the blog", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Blog posted!"))
	if err := json.NewEncoder(w).Encode(blog); err != nil {
		http.Error(w, "Unable to post blog at the moment", http.StatusInternalServerError)
		return
	}
}

func (h *BlogHandler) GetBlog (w http.ResponseWriter, r *http.Request) {
	var blogs = []models.Blog{}

	filter := bson.M{}
	cursor, err := h.Collection.Find(context.TODO(), filter)
	if err != nil {
		http.Error(w, "Error getting the blogs", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &blogs); err != nil {
		http.Error(w, "Error decoding blogs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(blogs); err != nil {
		http.Error(w, "Error encoding the blogs", http.StatusInternalServerError)
		return
	}
}