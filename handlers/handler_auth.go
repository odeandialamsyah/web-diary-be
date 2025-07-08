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
