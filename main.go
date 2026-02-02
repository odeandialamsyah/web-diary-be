package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors" // Untuk menangani CORS

	"web-diary-be/config"
	"web-diary-be/routes"
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
		AllowOrigins:     "http://localhost:60225",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders:    "Content-Length, Access-Control-Allow-Origin",
		AllowCredentials: false, // set to true only if frontend sends cookies/credentials
		MaxAge:           3600,
	}))
	routes.AuthRoutes(app) // Rute untuk otentikasi
	routes.DiaryRoutes(app)


	// Jalankan server
	port := ":8080"
	log.Printf("Server is running on port %s", port)
	log.Fatal(app.Listen(port))
}