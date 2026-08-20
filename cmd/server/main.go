package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"stockopname-rita-backend/internal/router"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r := router.NewRouter()

	host := os.Getenv("APP_HOST")
	port := os.Getenv("APP_PORT")

	addr := host + ":" + port

	fmt.Println("Server is running on", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
