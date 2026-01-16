package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"goth/internals/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthHandler struct {
	Collection *mongo.Collection
}

func (h *AuthHandler) Signup (w http.ResponseWriter, r *http.Request) {
	var user models.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()

	_, err := h.Collection.InsertOne(context.TODO(), user)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Signup successfull"))
	json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) Login (w http.ResponseWriter, r *http.Request) {
	var inputs models.User
	var foundUser models.User

	if err := json.NewDecoder(r.Body).Decode(&inputs); err != nil {
		http.Error(w, "Invaid Credentials", http.StatusUnauthorized)
		return
	}

	filter := bson.M{"email": inputs.Email}
	err := h.Collection.FindOne(context.TODO(), filter).Decode(&foundUser)

	if err != nil {
		http.Error(w, "User does not exists", http.StatusUnauthorized)
		return
	}

	if inputs.Password != foundUser.Password {
		http.Error(w, "Incorrect password", http.StatusUnauthorized)
		return
	} 

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Login success!"))
}

func (h *AuthHandler) GetUsers (w http.ResponseWriter, r *http.Request) {
	users := []models.User{}

	filter := bson.M{}
	
	cursor, err := h.Collection.Find(context.TODO(), filter)		// FIND returns a pointer to cursor
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &users); err != nil {
		http.Error(w, "Error decoding users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, "Error encoding users", http.StatusInternalServerError)
		return
	}
}