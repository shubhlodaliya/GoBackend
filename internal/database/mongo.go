package database

import (
	"backend/internal/config"
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

var UserCollection *mongo.Collection
var DeviceCollection *mongo.Collection

func ConnectDB() {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.Current.Database.Uri))
	if err != nil {
		log.Fatal("❌ MongoDB connection error:", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("❌ MongoDB ping failed:", err)
	}

	DB = client.Database("iotdb")

	UserCollection = DB.Collection("users")
	DeviceCollection = DB.Collection("devices")

	// Ensure index on mobile number for uniqueness
	_, err = UserCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"mobile": 1},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Fatalf("❌ Failed to create index: %v", err)
	}

	fmt.Println("✅ Connected to MongoDB!")
}
