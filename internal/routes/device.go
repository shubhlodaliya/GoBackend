package routes

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func UpdateDeviceStatus(w http.ResponseWriter, r *http.Request) {
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

func AddDevice(w http.ResponseWriter, r *http.Request) {
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
