package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

// Collector receives UDP events from game servers and publishes them to the message queue
type Collector struct {
	addr      string
	conn      *net.UDPConn
	publisher message.Publisher
	logger    watermill.LoggerAdapter
}

// Config holds collector configuration
type Config struct {
	UDPPort   int
	Publisher message.Publisher
	Logger    watermill.LoggerAdapter
}

// New creates a new collector
func New(cfg Config) *Collector {
	return &Collector{
		addr:      fmt.Sprintf(":%d", cfg.UDPPort),
		publisher: cfg.Publisher,
		logger:    cfg.Logger,
	}
}

// Start starts the UDP collector
func (c *Collector) Start(ctx context.Context) error {
	// Resolve UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", c.addr)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	// Listen for UDP packets
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP: %w", err)
	}
	c.conn = conn

	c.logger.Info("UDP collector started", watermill.LogFields{
		"address": c.addr,
	})

	// Start reading packets
	go c.readLoop(ctx)

	// Wait for context cancellation
	<-ctx.Done()
	return c.conn.Close()
}

// readLoop reads UDP packets and publishes them
func (c *Collector) readLoop(ctx context.Context) {
	buffer := make([]byte, 4096)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Set read deadline to allow context cancellation
		if err := c.conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
			c.logger.Error("Failed to set read deadline", err, watermill.LogFields{})
			continue
		}

		n, addr, err := c.conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			c.logger.Error("Failed to read UDP packet", err, watermill.LogFields{})
			continue
		}

		// Parse and publish event
		go c.handlePacket(buffer[:n], addr)
	}
}

// handlePacket processes a single UDP packet
func (c *Collector) handlePacket(data []byte, addr *net.UDPAddr) {
	// Parse JSON
	var rawEvent map[string]interface{}
	if err := json.Unmarshal(data, &rawEvent); err != nil {
		c.logger.Error("Failed to parse JSON", err, watermill.LogFields{
			"source": addr.String(),
			"data":   string(data),
		})
		return
	}

	// Validate required fields
	eventType, ok := rawEvent["event_type"].(string)
	if !ok {
		c.logger.Error("Missing event_type", nil, watermill.LogFields{
			"source": addr.String(),
		})
		return
	}

	// Create watermill message
	msg := message.NewMessage(watermill.NewUUID(), data)
	msg.Metadata.Set("event_type", eventType)
	msg.Metadata.Set("source_ip", addr.IP.String())
	msg.Metadata.Set("received_at", time.Now().Format(time.RFC3339))

	// Publish to appropriate topic based on event type
	topic := fmt.Sprintf("events.%s", eventType)

	if err := c.publisher.Publish(topic, msg); err != nil {
		c.logger.Error("Failed to publish message", err, watermill.LogFields{
			"topic":      topic,
			"event_type": eventType,
			"source":     addr.String(),
		})
		return
	}

	c.logger.Debug("Event published", watermill.LogFields{
		"topic":      topic,
		"event_type": eventType,
		"source":     addr.String(),
	})
}

// Stop stops the collector
func (c *Collector) Stop() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
