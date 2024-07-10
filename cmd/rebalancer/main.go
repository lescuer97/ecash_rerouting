package main

import (
	"fmt"
	"github.com/joho/godotenv"
	// "github.com/lescuer97/ecash_rerouting/internal/communication"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	// setup GRPC connection to LND

	fmt.Println("Hello, World!")
}
