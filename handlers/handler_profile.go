package handlers

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"web-diary-be/config"
	"web-diary-be/models"
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
		})
	}

	update := bson.M{}

	if payload.Username != nil {
		if *payload.Username == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "username cannot be empty",
			})
		}
		update["username"] = *payload.Username
	}

	if payload.Email != nil {
		if *payload.Email == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "email cannot be empty",
			})
		}

		// optional: cek email unik
		var existing models.User
		err := config.UserCollection.FindOne(
			context.Background(),
			bson.M{"email": *payload.Email, "_id": bson.M{"$ne": objID}},
		).Decode(&existing)
		if err == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "email already in use",
			})
		}

		update["email"] = *payload.Email
	}

	if payload.Password != nil {
		if len(*payload.Password) < 6 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "password must be at least 6 characters",
			})
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(*payload.Password), 14)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "failed to hash password",
			})
		}

		update["password"] = string(hashed)
	}

	if len(update) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "no updatable fields provided",
		})
	}

	update["updated_at"] = time.Now()

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var user models.User
	err = config.UserCollection.
		FindOneAndUpdate(
			context.Background(),
			bson.M{"_id": objID},
			bson.M{"$set": update},
			opts,
		).
		Decode(&user)

	if err != nil {
		log.Printf("Error updating profile: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to update profile",
		})
	}

	// response konsisten dengan Me
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
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

