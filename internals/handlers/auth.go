package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
	"strings"

	"goth/internals/helpers"

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

	// setting up default fields
	user.Role = "user"
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()

	_, err := h.Collection.InsertOne(context.TODO(), user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			if strings.Contains(err.Error(), "email") {
				http.Error(w, "User with email already exists", http.StatusConflict)
				return
			}
			if strings.Contains(err.Error(), "userName") {
				http.Error(w, "That username is already taken", http.StatusConflict)
				return
			}
		}
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{} {
		"success": "true",
		"message": "User created successfully!",
		"user": user,
	}
	
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding the response", http.StatusInternalServerError)
		return
	}
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

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, "Error encoding users", http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandler) Login (w http.ResponseWriter, r *http.Request) {
	var inputs models.User
	var foundUser models.User

	// decoding the req object
	if err := json.NewDecoder(r.Body).Decode(&inputs); err != nil {
		http.Error(w, "Error decoding the request", http.StatusInternalServerError)
		return
	}

	// searching if the email entered exists?
	filter := bson.M{"email": inputs.Email}
	err := h.Collection.FindOne(context.TODO(), filter).Decode(&foundUser)
	if err != nil {
		http.Error(w, "No user with that email exists", http.StatusBadRequest)
		return
	}

	// Validating the passwords
	if inputs.Password != foundUser.Password {
		http.Error(w, "Incorrect Password", http.StatusUnauthorized)
		return
	}

	tokenString, err := helpers.GenerateToken(foundUser.ID.Hex(), foundUser.Email);
	if err != nil {
		http.Error(w, "Error in generating a token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name: "token",
		Value: tokenString,
		Path: "/",
		HttpOnly: true,
		Secure: false,
		SameSite: http.SameSiteLaxMode,
		Expires: time.Now().Add(24 * time.Hour),
	})

	response := map[string]interface{} {
		"message": "success",
		"user": foundUser,
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding the response", http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandler) Logout (w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name: "token",
		Value: " ",
		Path: "/",
		Expires: time.Unix(0, 0),
		HttpOnly: true,
		MaxAge: -1,
	})

	response := map[string]interface{} {
		"status": "success",
		"message": "Logged out successfully!",
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding the response", http.StatusInternalServerError)
		return
	}
}
