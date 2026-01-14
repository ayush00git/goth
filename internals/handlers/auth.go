package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"goth/internals/models"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthHandler struct {
	Collection *mongo.Collection
}

func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var user models.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user.ID = primitive.NewObjectID();
	user.CreatedAt = time.Now();

	_, err := h.Collection.InsertOne(context.TODO(), user)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated) // 201
	w.Write([]byte("User created successfully"))
	json.NewEncoder(w).Encode(user);
}

func (h *AuthHandler) LogIn(w http.ResponseWriter, r *http.Request) {
	var inputs models.User
	var foundUser models.User

	if err := json.NewDecoder(r.Body).Decode(&inputs); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest);
		return;
	}

	filter := bson.M{"email": inputs.Email};
	err := h.Collection.FindOne(context.TODO(), filter).Decode(&foundUser);

	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized);
		return;
	}

	if inputs.Password != foundUser.Password {
		http.Error(w, "Password is incorrect", http.StatusUnauthorized);
		return;
	}
	w.WriteHeader(http.StatusAccepted);
	w.Write([]byte("Logged in success"));
}
