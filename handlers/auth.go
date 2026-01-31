package handlers

import (
	"context"
	"net/http"
	"time"
	"strings"

	"goth/helpers"
	"goth/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthHandler struct {
	Collection *mongo.Collection
}

func (h *AuthHandler) Signup (c *gin.Context) {
	// get json request
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// password hashing
	hashPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating hash"})
		return
	}

	// save defined fields
	user.Password = string(hashPass)
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.Role = "user"

	_, err = h.Collection.InsertOne(context.TODO(), user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			if strings.Contains(err.Error(), "email") {
				c.JSON(http.StatusConflict, gin.H{"error": "User with that email already exists"})
				return
			}
			if strings.Contains(err.Error(), "userName") {
				c.JSON(http.StatusConflict, gin.H{"error": "That username is already taken"})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving to database"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": "Signup success!",
		"user": map[string]interface{} {
			"userName": user.UserName,
			"email": user.Email,
			"role": user.Role,
		},
 	})
}

func (h *AuthHandler) GetUsersGin (c *gin.Context) {
	var users = []models.User{}

	// fetch all users
	cursor, err := h.Collection.Find(context.TODO(), bson.M{})		// .Find returns a cursor pointer
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fetching users from mongodb"})
		return
	}
	defer cursor.Close(context.TODO())

	// We can use the cursor.All() or cursor.Next() to iterate through the documents
	if err := cursor.All(context.TODO(), &users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to iterate through at the moment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": "Users fetched successfully",
		"users": users,
	})
}

func (h *AuthHandler) Login (c *gin.Context) {
	// get the json request
	var user models.User
	var foundUser models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// finding the user
	filter := bson.M{"email": user.Email}
	err := h.Collection.FindOne(context.TODO(), filter).Decode(&foundUser)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No user found"})
		return
	}
	
	// comparing the hash and input password
	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(user.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect password"})
		return
	}

	// generate a JWT token
	tokenString, err := helpers.GenerateToken(foundUser.ID.Hex(), foundUser.Email, foundUser.UserName, foundUser.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "generating a jwt token"})
		return
	}

	c.SetCookie("token", tokenString, 3600, "/", "localhost", false, true)		// c.SetCookie(Name, Value, MaxAge, Path, Domain, Secure, HttpOnly)
	c.JSON(http.StatusOK, gin.H{"success": "Logged In successfully!"})
}

func (h *AuthHandler) Logout (c *gin.Context) {
	c.SetCookie("token", " ", -1, "/", "localhost", false, true)
	c.JSON(http.StatusOK, gin.H{"success": "Logged out successfully!"})
}
