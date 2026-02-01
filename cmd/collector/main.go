package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/UDL-TF/UnitedStats/internal/collector"
	"github.com/UDL-TF/UnitedStats/internal/queue"
)

func main() {
	// Create logger
	logger := watermill.NewStdLogger(false, false)

	// Get configuration from environment
	rabbitmqURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	udpPort := getEnvInt("UDP_PORT", 27500)

	// Create RabbitMQ publisher
	publisher, err := queue.NewPublisher(queue.Config{
		URL:    rabbitmqURL,
		Logger: logger,
	})
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}
	defer publisher.Close()

	// Create collector
	c := collector.New(collector.Config{
		UDPPort:   udpPort,
		Publisher: publisher,
		Logger:    logger,
	})

	// Start collector
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down collector...")
		cancel()
	}()

	log.Printf("Starting collector on UDP port %d...\n", udpPort)
	if err := c.Start(ctx); err != nil {
		log.Fatalf("Collector error: %v", err)
	}

	log.Println("Collector stopped")
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
