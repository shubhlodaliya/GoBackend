package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Device struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID   primitive.ObjectID `bson:"user_id" json:"user_id"`
	DeviceID string             `bson:"device_id" json:"device_id"`
	Name     string             `bson:"name" json:"name"`
	Type     string             `bson:"type" json:"type"`     // Always "static"
	Status   string             `bson:"status" json:"status"` // "on" or "off"
}
