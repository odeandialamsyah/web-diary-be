package routes

import (
	"web-diary-be/handlers"
	"web-diary-be/middleware"

	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(app *fiber.App) {
	api := app.Group("/api/auth")

	api.Post("/register", handlers.Register)
	api.Post("/login", handlers.Login)
	api.Get("/logout", handlers.Logout) // logout bisa di-handle client-side
}

func DiaryRoutes(app *fiber.App) {
	diary := app.Group("/api/diary")
	diary.Use(middleware.JWTProtected()) // Wajibkan token

	diary.Post("/", handlers.CreateDiaryEntry)
	diary.Get("/", handlers.GetDiaryEntries)
	diary.Get("/:id", handlers.GetDiaryEntryByID)	
	diary.Put("/:id", handlers.UpdateDiaryEntry)
	diary.Delete("/:id", handlers.DeleteDiaryEntry)
}

func ProfileRoutes(app *fiber.App) {
	profile := app.Group("/api/profile")

	profile.Use(middleware.JWTProtected())

	profile.Get("/me", handlers.Me)
	profile.Put("/:id", handlers.UpdateProfile)
	profile.Delete("/:id", handlers.DeleteProfile)
}