package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"

	"stockopname-rita-backend/internal/database"
	"stockopname-rita-backend/internal/middleware"
	"stockopname-rita-backend/internal/router"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	host := os.Getenv("APP_HOST")
	port := os.Getenv("APP_PORT")
	databaseURL := os.Getenv("DATABASE_URL")

	addr := host + ":" + port

	ctx := context.Background()
	validate := validator.New()

	pool, err := database.NewPostgresPool(ctx, databaseURL)
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
	defer pool.Close()

	r := router.NewRouter(validate, pool)

	middleware := middleware.CORS(r)

	fmt.Println("Server is running on", addr)
	log.Fatal(http.ListenAndServe(addr, middleware))
}
