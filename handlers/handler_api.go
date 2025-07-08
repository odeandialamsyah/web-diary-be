package handlers

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"web-diary-be/config"
	models "web-diary-be/models"
	"web-diary-be/services"
)

// CreateDiaryEntry membuat entri diary baru dengan analisis emosi
func CreateDiaryEntry(c *fiber.Ctx) error {
	entry := new(models.DiaryEntry)

	if err := c.BodyParser(entry); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	if entry.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Diary content cannot be empty",
		})
	}

	// Dapatkan user ID dari token
	val := c.Locals("user_id")
	userID, ok := val.(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid or missing token",
		})
	}
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid user ID format",
			"error":   err.Error(),
		})
	}
	entry.UserID = userObjID

	// Analisis emosi
	emotion, sentiment, err := services.AnalyzeEmotion(entry.Content)
	if err != nil {
		log.Printf("Failed to analyze emotion: %v", err)
		entry.Emotion = "Unknown"
		entry.Sentiment = "Neutral"
	} else {
		entry.Emotion = emotion
		entry.Sentiment = sentiment
	}

	entry.ID = primitive.NewObjectID()
	entry.CreatedAt = time.Now()

	_, err = config.DiaryCollection.InsertOne(context.Background(), entry)
	if err != nil {
		log.Printf("Error inserting diary entry: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create diary entry",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(entry)
}

func GetDiaryEntries(c *fiber.Ctx) error {
	// Ambil user_id dari JWT (disimpan oleh middleware di Locals)
	val := c.Locals("user_id")
	userID, ok := val.(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or missing token",
		})
	}

	// Konversi string ke ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println("❌ Invalid user_id format:", userID)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	// // Log debug (opsional)
	// log.Println("✅ user_id (string):", userID)
	// log.Println("✅ user_id (ObjectID):", objID)

	// Persiapkan query dan sorting
	filter := bson.M{"user_id": objID}
	findOptions := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	// Query ke database
	cursor, err := config.DiaryCollection.Find(context.Background(), filter, findOptions)
	if err != nil {
		log.Printf("Error finding diary entries: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to retrieve diary entries",
			"error":   err.Error(),
		})
	}
	defer cursor.Close(context.Background())

	// Iterasi hasil
	var entries []models.DiaryEntry
	for cursor.Next(context.Background()) {
		var entry models.DiaryEntry
		if err := cursor.Decode(&entry); err != nil {
			log.Printf("Error decoding diary entry: %v", err)
			continue
		}
		entries = append(entries, entry)
	}

	// Cek jika ada error di cursor
	if err := cursor.Err(); err != nil {
		log.Printf("Cursor error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error while processing diary entries",
			"error":   err.Error(),
		})
	}

	// Kembalikan hasil
	return c.Status(fiber.StatusOK).JSON(entries)
}


// GetDiaryEntryByID mengambil satu entri diary berdasarkan ID
func GetDiaryEntryByID(c *fiber.Ctx) error {
	val := c.Locals("user_id")
	userID, ok := val.(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid or missing token",
		})
	}

	idParam := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid ID format",
		})
	}

	var entry models.DiaryEntry
	filter := bson.M{"_id": objID, "user_id": userID}
	err = config.DiaryCollection.FindOne(context.Background(), filter).Decode(&entry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Diary entry not found or not authorized",
			})
		}
		log.Printf("Error finding diary entry by ID: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to retrieve diary entry",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(entry)
}
