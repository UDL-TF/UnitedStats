package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/UDL-TF/UnitedStats/internal/processor"
	"github.com/UDL-TF/UnitedStats/internal/queue"
	"github.com/UDL-TF/UnitedStats/internal/store"
)

func main() {
	// Create logger
	logger := watermill.NewStdLogger(false, false)

	// Get configuration from environment
	rabbitmqURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnvInt("DB_PORT", 5432)
	dbUser := getEnv("DB_USER", "unitedstats")
	dbPassword := getEnv("DB_PASSWORD", "unitedstats")
	dbName := getEnv("DB_NAME", "unitedstats")

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
	defer st.Close()

	log.Println("Database connection established")

	// Create RabbitMQ subscriber
	subscriber, err := queue.NewSubscriber(queue.Config{
		URL:    rabbitmqURL,
		Logger: logger,
	})
	if err != nil {
		log.Fatalf("Failed to create subscriber: %v", err)
	}
	defer subscriber.Close()

	// Create processor
	proc := processor.New(processor.Config{
		Store:      st,
		Subscriber: subscriber,
		Logger:     logger,
	})

	// Start processor
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down processor...")
		cancel()
	}()

	log.Println("Starting event processor...")
	if err := proc.Start(ctx); err != nil {
		log.Fatalf("Processor error: %v", err)
	}

	log.Println("Processor stopped")
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
