package routes

import (
	"backend/internal/database"
	"backend/internal/models"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// POST /auth
func AuthHandler(c *gin.Context) {
	var req struct {
		Mobile string `json:"mobile"`
		Token  string `json:"token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.Mobile == "" || req.Token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	filter := bson.M{"mobile": req.Mobile}
	var user models.User
	err := database.UserCollection.FindOne(context.TODO(), filter).Decode(&user)

	if err == mongo.ErrNoDocuments {
		// New user: insert
		user = models.User{
			Mobile: req.Mobile,
			Token:  req.Token,
		}
		result, err := database.UserCollection.InsertOne(context.TODO(), user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User creation failed"})
			return
		}

		if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
			user.ID = oid
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse inserted ID"})
			return
		}
	} else if err == nil {
		// Existing user: update token
		update := bson.M{"$set": bson.M{"token": req.Token}}
		_, err := database.UserCollection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token update failed"})
			return
		}

		err = database.UserCollection.FindOne(context.TODO(), filter).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Fetch after update failed"})
			return
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_id": user.ID.Hex()})
}
