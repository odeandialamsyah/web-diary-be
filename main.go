package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors" // Untuk menangani CORS

	"web-diary-be/config"
	"web-diary-be/handlers"
)

func main() {
	// Load environment variables dari .env
	config.LoadEnv()

	// Koneksi ke MongoDB
	config.ConnectDB()
	defer config.DisconnectDB() // Pastikan koneksi ditutup saat aplikasi berhenti

	app := fiber.New()

	// Middleware CORS agar frontend bisa mengakses API ini
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // Ganti dengan domain frontend Anda di produksi
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Routes
	app.Post("/api/diary-entries", handlers.CreateDiaryEntry)
	app.Get("/api/diary-entries", handlers.GetDiaryEntries)
	app.Get("/api/diary-entries/:id", handlers.GetDiaryEntryByID)


	// Jalankan server
	port := ":3000" // Atau dari environment variable
	log.Printf("Server is running on port %s", port)
	log.Fatal(app.Listen(port))
}