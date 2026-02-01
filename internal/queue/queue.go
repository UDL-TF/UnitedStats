package queue

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v2/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
)

// Config holds RabbitMQ configuration
type Config struct {
	URL    string // amqp://user:pass@localhost:5672/
	Logger watermill.LoggerAdapter
}

// NewPublisher creates a new RabbitMQ publisher
func NewPublisher(cfg Config) (message.Publisher, error) {
	amqpConfig := amqp.NewDurablePubSubConfig(cfg.URL, nil)

	publisher, err := amqp.NewPublisher(amqpConfig, cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher: %w", err)
	}

	return publisher, nil
}

// NewSubscriber creates a new RabbitMQ subscriber
func NewSubscriber(cfg Config) (message.Subscriber, error) {
	amqpConfig := amqp.NewDurablePubSubConfig(cfg.URL, nil)

	subscriber, err := amqp.NewSubscriber(amqpConfig, cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscriber: %w", err)
	}

	return subscriber, nil
}
