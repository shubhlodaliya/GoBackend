package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

// ----------------------
// MongoDB Setup
// ----------------------

var client *mongo.Client
var userCollection *mongo.Collection

func connectDB() {
    uri := "mongodb+srv://farmerIOT:FarmerIOT@cluster0.phbob3u.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0" // 🔁 Replace with your real MongoDB URI
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var err error
    client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
    if err != nil {
        log.Fatalf("❌ MongoDB connection error: %v", err)
    }

    // Ping test
    err = client.Ping(ctx, nil)
    if err != nil {
        log.Fatalf("❌ MongoDB ping error: %v", err)
    }

    db := client.Database("your_db_name") // 🔁 Replace with your database name
    userCollection = db.Collection("users")

    // Ensure index on mobile number for uniqueness
    _, err = userCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
        Keys:    bson.M{"mobile": 1},
        Options: options.Index().SetUnique(true),
    })
    if err != nil {
        log.Fatalf("❌ Failed to create index: %v", err)
    }

    fmt.Println("✅ Connected to MongoDB and user collection is ready!")
}

// ----------------------
// Model
// ----------------------

type User struct {
    ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Mobile string             `bson:"mobile" json:"mobile"`
    Token  string             `bson:"token" json:"token"`
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

    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil || req.Mobile == "" || req.Token == "" {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    var user User
    err = userCollection.FindOne(context.TODO(), bson.M{"mobile": req.Mobile}).Decode(&user)

    if err == mongo.ErrNoDocuments {
        // Create new user
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
    }

    json.NewEncoder(w).Encode(map[string]string{
        "user_id": user.ID.Hex(),
    })
}

// ----------------------
// Main
// ----------------------

func main() {
    connectDB()

    r := mux.NewRouter()
    r.HandleFunc("/api/auth/login", AuthHandler).Methods("POST")

    fmt.Println("🚀 Server running at http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}
