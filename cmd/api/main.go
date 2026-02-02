package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/UDL-TF/UnitedStats/internal/api"
	"github.com/UDL-TF/UnitedStats/internal/store"
)

func main() {
	// Get configuration from environment
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnvInt("DB_PORT", 5432)
	dbUser := getEnv("DB_USER", "unitedstats")
	dbPassword := getEnv("DB_PASSWORD", "unitedstats")
	dbName := getEnv("DB_NAME", "unitedstats")
	apiPort := getEnvInt("API_PORT", 8080)

	// Create database store
	st, err := store.New(store.Config{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		DBName:   dbName,
		SSLMode:  "disable",
	})
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer func() {
		if err := st.Close(); err != nil {
			log.Printf("Error closing store: %v", err)
		}
	}()

	log.Println("Database connection established")

	// Create API server
	server := api.New(api.Config{
		Store: st,
		Port:  apiPort,
	})

	// Handle shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		log.Printf("Starting API server on port %d...\n", apiPort)
		if err := server.Start(); err != nil {
			log.Printf("API server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down API server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("API server shutdown error: %v", err)
	}

	log.Println("API server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}
