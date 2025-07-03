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
	models "web-diary-be/model"
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

	// Validasi input minimal
	if entry.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Diary content cannot be empty",
		})
	}

	// Panggil Gemini Flash API untuk menganalisis emosi
	emotion, sentiment, err := services.AnalyzeEmotion(entry.Content)
	if err != nil {
		log.Printf("Failed to analyze emotion: %v", err)
		// Tetap simpan entri meskipun analisis gagal, mungkin dengan emosi default
		entry.Emotion = "Unknown"
		entry.Sentiment = "Neutral"
	} else {
		entry.Emotion = emotion
		entry.Sentiment = sentiment
	}

	entry.ID = primitive.NewObjectID()
	entry.CreatedAt = time.Now()

	_, err = config.Collection.InsertOne(context.Background(), entry)
	if err != nil {
		log.Printf("Error inserting diary entry: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create diary entry",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(entry)
}

// GetDiaryEntries mengambil semua entri diary
func GetDiaryEntries(c *fiber.Ctx) error {
	var entries []models.DiaryEntry
	findOptions := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}) // Urutkan dari terbaru

	cursor, err := config.Collection.Find(context.Background(), bson.M{}, findOptions)
	if err != nil {
		log.Printf("Error finding diary entries: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to retrieve diary entries",
			"error":   err.Error(),
		})
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var entry models.DiaryEntry
		cursor.Decode(&entry)
		entries = append(entries, entry)
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Error iterating cursor: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error processing diary entries",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(entries)
}

// GetDiaryEntryByID mengambil satu entri diary berdasarkan ID
func GetDiaryEntryByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid ID format",
		})
	}

	var entry models.DiaryEntry
	err = config.Collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&entry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Diary entry not found",
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