package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"mussinalin/interview_bedrock/database"

	"github.com/joho/godotenv"
)

func main() {
	// load .env
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v\n", err)
	}

	// read database dsn
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)

	// init database
	if err := database.InitDB(dsn); err != nil {
		log.Fatalf("Failed to initialize database: %v\n", err)
	}
	defer database.CloseDB()

	ctx := context.Background()

	// test insert
	if err := database.InsertTx(ctx, "0x1234", 999, "0x2222", "0xmussinaeth"); err != nil {
		log.Fatalf("Error inserting user: %v\n", err)
	}
}
