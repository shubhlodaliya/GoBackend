package routes

import (
	"context"
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AuthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Mobile string `json:"mobile"`
		Token  string `json:"token"`
	}
	// Decode request body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Mobile == "" || req.Token == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Try to find the user by mobile
	filter := bson.M{"mobile": req.Mobile}
	var user User
	err = userCollection.FindOne(context.TODO(), filter).Decode(&user)

	if err == mongo.ErrNoDocuments {
		// 🔵 New user: insert
		user = User{
			Mobile: req.Mobile,
			Token:  req.Token,
		}
		result, err := userCollection.InsertOne(context.TODO(), user)
		if err != nil {
			http.Error(w, "User creation failed", http.StatusInternalServerError)
			return
		}
		user.ID = result.InsertedID.(primitive.ObjectID)
	} else if err == nil {
		// 🟠 Existing user: update token
		update := bson.M{"$set": bson.M{"token": req.Token}}
		_, err := userCollection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			http.Error(w, "Token update failed", http.StatusInternalServerError)
			return
		}

		// ⚠️ Fetch updated user
		err = userCollection.FindOne(context.TODO(), filter).Decode(&user)
		if err != nil {
			http.Error(w, "Fetch after update failed", http.StatusInternalServerError)
			return
		}
	} else {
		// 🔴 Unexpected error
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// ✅ Return user ID
	json.NewEncoder(w).Encode(map[string]string{
		"user_id": user.ID.Hex(),
	})
}
