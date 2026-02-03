package handlers

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"web-diary-be/config"
	"web-diary-be/models"
	"web-diary-be/services"
)

// UpdateMe memperbarui profil user yang sedang login
func UpdateProfile(c *fiber.Ctx) error {
	val := c.Locals("user_id")
	userID, ok := val.(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid or missing token",
		})
	}

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id format",
		})
	}

	// payload partial update
	var payload struct {
		Username *string `json:"username"`
		Email    *string `json:"email"`
		Password *string `json:"password"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
			"detail": err.Error(),
		})
	}

	setFields := bson.M{}

	if payload.Username != nil {
		if *payload.Username == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "username cannot be empty",
			})
		}
		setFields["username"] = *payload.Username
	}

	if payload.Email != nil {
		if *payload.Email == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "email cannot be empty",
			})
		}
		setFields["email"] = *payload.Email
	}

	if payload.Password != nil {
		if len(*payload.Password) < 6 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "password must be at least 6 characters",
			})
		}
		// ⚠️ pastikan ini di-hash
		hashed, err := services.HashPassword(*payload.Password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "failed to hash password",
			})
		}
		setFields["password"] = hashed
	}

	if len(setFields) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "no updatable fields provided",
		})
	}

	setFields["updated_at"] = time.Now()

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated models.User
	err = config.UserCollection.
		FindOneAndUpdate(
			context.Background(),
			bson.M{"_id": objID},
			bson.M{"$set": setFields},
			opts,
		).
		Decode(&updated)

	if err != nil {
		log.Printf("Error updating user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to update profile",
		})
	}

	// response tanpa password
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"id":         updated.ID,
		"username":   updated.Username,
		"email":      updated.Email,
		"created_at": updated.CreatedAt,
		"updated_at": updated.UpdatedAt,
	})
}

// DeleteMe menghapus akun user yang sedang login
func DeleteProfile(c *fiber.Ctx) error {
	val := c.Locals("user_id")
	userID, ok := val.(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid or missing token",
		})
	}

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id format",
		})
	}

	// (opsional) hapus semua diary user
	_, err = config.DiaryCollection.DeleteMany(
		context.Background(),
		bson.M{"user_id": objID},
	)
	if err != nil {
		log.Printf("Failed deleting user diaries: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to delete user diaries",
		})
	}

	// hapus user
	res, err := config.UserCollection.DeleteOne(
		context.Background(),
		bson.M{"_id": objID},
	)
	if err != nil {
		log.Printf("Failed deleting user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to delete account",
		})
	}

	if res.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "user not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "account deleted successfully",
	})
}

