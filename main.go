package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"backend/internal/config"
	"backend/internal/database"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type User struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Mobile string             `bson:"mobile" json:"mobile"`
	Token  string             `bson:"token" json:"token"`
}

type Device struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"` // MongoDB internal ID
	UserID   primitive.ObjectID `bson:"user_id" json:"user_id"`
	DeviceID string             `bson:"device_id" json:"device_id"` // 👈 Custom device ID from request
	Name     string             `bson:"name" json:"name"`
	Type     string             `bson:"type" json:"type"`     // always "static"
	Status   string             `bson:"status" json:"status"` // "on" or "off"
}

// ----------------------
// Handler
// ----------------------

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
func AddDeviceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		UserID   string `json:"user_id"`
		DeviceID string `json:"device_id"` // 👈 New field
		Name     string `json:"name"`
	}

	// Validate input
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.UserID == "" || req.DeviceID == "" || req.Name == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	userID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Prepare new device object
	newDevice := Device{
		UserID:   userID,
		DeviceID: req.DeviceID, // 👈 Set from request
		Name:     req.Name,
		Type:     "static",
		Status:   "off", // default
	}

	// Insert into DB
	_, err = deviceCollection.InsertOne(context.TODO(), newDevice)
	if err != nil {
		http.Error(w, "Failed to add device", http.StatusInternalServerError)
		return
	}

	// Fetch all devices for this user
	cursor, err := deviceCollection.Find(context.TODO(), bson.M{"user_id": userID})
	if err != nil {
		http.Error(w, "Failed to fetch devices", http.StatusInternalServerError)
		return
	}

	var devices []Device
	if err = cursor.All(context.TODO(), &devices); err != nil {
		http.Error(w, "Cursor decode error", http.StatusInternalServerError)
		return
	}

	// Return device list
	json.NewEncoder(w).Encode(devices)
}

func UpdateDeviceStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		DeviceID string `json:"device_id"`
		Status   string `json:"status"` // "on" or "off"
	}

	// Validate input
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.DeviceID == "" || (req.Status != "on" && req.Status != "off") {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Update using custom device_id (not _id)
	filter := bson.M{"device_id": req.DeviceID}
	update := bson.M{"$set": bson.M{"status": req.Status}}

	result, err := deviceCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		http.Error(w, "Update failed", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	// Respond with success
	json.NewEncoder(w).Encode(map[string]string{"message": "Status updated successfully"})
}

func GetDevicesByUserID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userIDStr := mux.Vars(r)["user_id"]
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Find devices by user ID
	filter := bson.M{"user_id": userID}
	cursor, err := deviceCollection.Find(context.TODO(), filter)
	if err != nil {
		http.Error(w, "Failed to fetch devices", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var devices []Device
	for cursor.Next(context.TODO()) {
		var device Device
		if err := cursor.Decode(&device); err != nil {
			http.Error(w, "Decode error", http.StatusInternalServerError)
			return
		}
		devices = append(devices, device)
	}

	json.NewEncoder(w).Encode(devices)
}

// ----------------------
// Main
// ----------------------

func Main() {
	err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config file: %s", err)
	}
	database.ConnectDB()
	r := mux.NewRouter()
	r.HandleFunc("/api/auth/login", AuthHandler).Methods("POST")
	r.HandleFunc("/api/device/add", AddDeviceHandler).Methods("POST")
	r.HandleFunc("/api/device/status", UpdateDeviceStatusHandler).Methods("POST")
	r.HandleFunc("/api/devices/{user_id}", GetDevicesByUserID).Methods("GET")

	fmt.Println("🚀 Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
