package backend

import (
	"log"

	"backend/internal/config"
	"backend/internal/database"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Device struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"` // MongoDB internal ID
	UserID   primitive.ObjectID `bson:"user_id" json:"user_id"`
	DeviceID string             `bson:"device_id" json:"device_id"` // 👈 Custom device ID from request
	Name     string             `bson:"name" json:"name"`
	Type     string             `bson:"type" json:"type"`     // always "static"
	Status   string             `bson:"status" json:"status"` // "on" or "off"
}

func Main() {
	err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config file: %s", err)
	}
	database.ConnectDB()
	run()
}

func run() {
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	r := gin.Default()

	// routes.Init(r)

	r.Run(config.Current.Port)
	log.Println("Starting server on", config.Current.Port)
}
