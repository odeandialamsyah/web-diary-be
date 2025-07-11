package config

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client             *mongo.Client
	Database           *mongo.Database
	DiaryCollection    *mongo.Collection
	UserCollection     *mongo.Collection
	GeminiFlashAPIKey  string
)

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	GeminiFlashAPIKey = os.Getenv("GEMINI_FLASH_API_KEY")
	if GeminiFlashAPIKey == "" {
		log.Fatal("GEMINI_FLASH_API_KEY not set in .env")
	}
}

func ConnectDB() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI not set in .env")
	}

	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		log.Fatal("MONGO_DB_NAME not set in .env")
	}

	clientOptions := options.Client().ApplyURI(mongoURI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("✅ Connected to MongoDB!")
	Client = client
	Database = client.Database(dbName)

	DiaryCollection = Database.Collection("diary_entries")
	UserCollection = Database.Collection("users")
}

func DisconnectDB() {
	if Client == nil {
		return
	}
	err := Client.Disconnect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("🔌 Disconnected from MongoDB.")
}
