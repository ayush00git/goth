package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"goth/internals/models"
	"goth/internals/middlewares"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
)

type BlogHandler struct {
	Collection *mongo.Collection
}

func (h *BlogHandler) WriteBlog (w http.ResponseWriter, r *http.Request) {
	
	userIDStr, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}

	authorObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Error converting string to objId", http.StatusInternalServerError)
		return
	}

	var blog models.Blog

	blog.ID = primitive.NewObjectID()
	blog.AuthorID = authorObjID
	blog.CreatedAt = time.Now()

	if err := json.NewDecoder(r.Body).Decode(&blog); err != nil {
		http.Error(w, "Error fetching user's request", http.StatusBadRequest)
		return
	}

	_, err = h.Collection.InsertOne(context.TODO(), blog)
	if err != nil {
		http.Error(w, "Error submitting the blog", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Blog posted successfully",
		"blog": map[string]interface{}{
			"author": blog.AuthorID,
			"title": blog.Title,
			"excerpt": blog.Excerpt,
			"tags": blog.Tags,
			"isDraft": blog.IsDraft,
			"content": blog.Content,
		},
	}	
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Unable to post blog at the moment", http.StatusInternalServerError)
		return
	}
}

// Only get public blogs
func (h *BlogHandler) GetBlog (w http.ResponseWriter, r *http.Request) {
	var blogs = []models.Blog{}

	filter := bson.M{"isDraft": false}
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

func (h *BlogHandler) GetBlogByID (w http.ResponseWriter, r *http.Request) {
	ObjID := r.PathValue("BlogID")

	blogId, err := primitive.ObjectIDFromHex(ObjID)
	if err != nil {
		http.Error(w, "Error converting object id to string at getapi", http.StatusBadRequest)
		return
	}

	var foundBlog models.Blog

	filter := bson.M{"_id": blogId}
	err = h.Collection.FindOne(context.TODO(), filter).Decode(&foundBlog)

	if err != nil {
		http.Error(w, "No document with that ID exists", http.StatusBadRequest)
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(foundBlog); err != nil {
		http.Error(w, "Unable to encode the blog", http.StatusInternalServerError)
		return
	}
}

func (h *BlogHandler) DeleteBlogByID (w http.ResponseWriter, r *http.Request) {
	
	userIDStr, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(w, "Error finding user context", http.StatusUnauthorized)
		return
	}
	
	userIDObj, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Error converting from string to objId(user)", http.StatusInternalServerError)
		return
	}

	blogIDStr := r.PathValue("BlogID")

	blogIDObj, err := primitive.ObjectIDFromHex(blogIDStr)
	if err != nil {
		http.Error(w, "Error converting from string to objId(blog)", http.StatusInternalServerError)
		return
	}

	var deletedBlog models.Blog

	filter := bson.M{"_id": blogIDObj}
	err = h.Collection.FindOneAndDelete(context.TODO(), filter).Decode(&deletedBlog)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "No blogs found", http.StatusNotFound)
			return
		}
		http.Error(w, "Unable to find the blog", http.StatusBadRequest)
		return
	}
	
	if deletedBlog.AuthorID != userIDObj {
		http.Error(w, "You are not authorized for this action", http.StatusUnauthorized)
		return
	}

	response := map[string]interface{} {
		"success": "true",
		"message": "Blog deleted successfully!",
		"blog": deletedBlog,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Unable to encode json", http.StatusInternalServerError)
		return
	}
}

func (h *BlogHandler) EditBlogByID (w http.ResponseWriter, r *http.Request) {
	objId := r.PathValue("BlogID")

	blogId, err := primitive.ObjectIDFromHex(objId)
	if err != nil {
		http.Error(w, "Unable to convert from string to object id", http.StatusInternalServerError)
		return
	}

	var updatedData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updatedData); err != nil {
		http.Error(w, "Unable to edit", http.StatusInternalServerError)
		return
	}

	update := bson.M{"$set": updatedData}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedBlog models.Blog
	err = h.Collection.FindOneAndUpdate(context.TODO(), bson.M{"_id": blogId}, update, opts).Decode(&updatedBlog)
	if err != nil {
		http.Error(w, "Error updating the document", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Blog updated successfully",
		"blog": updatedBlog,
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "{error: encoding the json}", http.StatusInternalServerError)
		return
	}
}
