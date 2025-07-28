package routes

import (
	"backend/internal/database"
	"backend/internal/models"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// POST /device/status
func UpdateDeviceStatus(c *gin.Context) {
	var req struct {
		DeviceID string `json:"device_id"`
		Status   string `json:"status"` // "on" or "off"
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.DeviceID == "" || (req.Status != "on" && req.Status != "off") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	filter := bson.M{"device_id": req.DeviceID}
	update := bson.M{"$set": bson.M{"status": req.Status}}

	result, err := database.DeviceCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully"})
}

// POST /device
func AddDevice(c *gin.Context) {
	var req struct {
		UserID   string `json:"user_id"`
		DeviceID string `json:"device_id"`
		Name     string `json:"name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.UserID == "" || req.DeviceID == "" || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	newDevice := models.Device{
		UserID:   userID,
		DeviceID: req.DeviceID,
		Name:     req.Name,
		Type:     "static",
		Status:   "off",
	}

	_, err = database.DeviceCollection.InsertOne(context.TODO(), newDevice)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add device"})
		return
	}

	cursor, err := database.DeviceCollection.Find(context.TODO(), bson.M{"user_id": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch devices"})
		return
	}
	defer cursor.Close(context.TODO())

	var devices []models.Device
	if err := cursor.All(context.TODO(), &devices); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cursor decode error"})
		return
	}

	c.JSON(http.StatusOK, devices)
}

// GET /devices/:user_id
func GetDevicesByUserID(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	cursor, err := database.DeviceCollection.Find(context.TODO(), bson.M{"user_id": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch devices"})
		return
	}
	defer cursor.Close(context.TODO())

	var devices []models.Device
	if err := cursor.All(context.TODO(), &devices); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Decode error"})
		return
	}

	c.JSON(http.StatusOK, devices)
}
