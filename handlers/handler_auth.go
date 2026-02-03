package handlers

import (
	"context"
	"web-diary-be/config"
	"web-diary-be/middleware"
	models "web-diary-be/models"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *fiber.Ctx) error {
    collection := config.UserCollection

    var user models.User
    if err := c.BodyParser(&user); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
    }

    var existing models.User
    err := collection.FindOne(context.TODO(), bson.M{"email": user.Email}).Decode(&existing)
    if err == nil {
        return c.Status(400).JSON(fiber.Map{"error": "Email already exists"})
    }

    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), 14)
    user.Password = string(hashedPassword)

    _, err = collection.InsertOne(context.TODO(), user)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Register failed"})
    }

    return c.JSON(fiber.Map{"message": "Registration successful"})
}

func Login(c *fiber.Ctx) error {
    collection := config.UserCollection

    var input models.User
    if err := c.BodyParser(&input); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
    }

    var user models.User
    err := collection.FindOne(context.TODO(), bson.M{"email": input.Email}).Decode(&user)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Email not found"})
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Wrong password"})
    }

    // Gunakan GenerateJWT agar konsisten
    t, err := middleware.GenerateJWT(user.ID.Hex())
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Token creation failed"})
    }
    return c.JSON(fiber.Map{"token": t})
}

func Logout(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"message": "Logout success (client should delete token)"})
}

// Me mengembalikan profil user yang sedang login
func Me(c *fiber.Ctx) error {
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

	var user models.User
	err = config.UserCollection.
		FindOne(context.Background(), bson.M{"_id": objID}).
		Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "User not found",
			})
		}
		log.Printf("Error fetching user profile: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to fetch user profile",
			"error":   err.Error(),
		})
	}

	// Response tanpa password
	type UserResponse struct {
		ID        primitive.ObjectID `json:"id"`
		Username  string             `json:"username"`
		Email     string             `json:"email"`
		CreatedAt time.Time          `json:"created_at"`
		UpdatedAt time.Time          `json:"updated_at,omitempty"`
	}

	resp := UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}
