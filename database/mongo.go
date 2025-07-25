package database

import (
    "context"
    "fmt"
    "log"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

func ConnectDB() {
    uri := "mongodb+srv://farmerIOT:FarmerIOT@cluster0.phbob3u.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"
    clientOptions := options.Client().ApplyURI(uri)

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Fatal("❌ MongoDB connection error:", err)
    }

    err = client.Ping(ctx, nil)
    if err != nil {
        log.Fatal("❌ MongoDB ping failed:", err)
    }

    DB = client.Database("iotdb")
    fmt.Println("✅ Connected to MongoDB!")
}

func GetCollection(name string) *mongo.Collection {
    return DB.Collection(name)
}
