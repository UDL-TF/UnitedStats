package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/UDL-TF/UnitedStats/internal/parser"
	"github.com/UDL-TF/UnitedStats/internal/store"
	"github.com/UDL-TF/UnitedStats/pkg/events"
)

// Processor consumes events from the message queue and stores them in PostgreSQL
type Processor struct {
	store      *store.Store
	subscriber message.Subscriber
	logger     watermill.LoggerAdapter
}

// Config holds processor configuration
type Config struct {
	Store      *store.Store
	Subscriber message.Subscriber
	Logger     watermill.LoggerAdapter
}

// New creates a new processor
func New(cfg Config) *Processor {
	return &Processor{
		store:      cfg.Store,
		subscriber: cfg.Subscriber,
		logger:     cfg.Logger,
	}
}

// Start starts processing events
func (p *Processor) Start(ctx context.Context) error {
	// Subscribe to all event topics
	eventTypes := []string{
		"kill", "airshot", "deflect", "stun", "jarate", "shield_blocked",
		"rocket_jump", "sticky_jump", "rocket_jump_kill", "sticky_jump_kill",
		"teleport", "teleport_used",
		"built_object", "killed_object",
		"healed", "uber_deployed", "uber_dropped", "defended_medic",
		"buff_deployed", "sandvich", "dalokohs", "steak",
		"match_start", "match_end", "round_start", "round_end",
		"mvp1", "mvp2", "mvp3",
		"player_loadout", "weapon_stats", "class_change",
	}

	for _, eventType := range eventTypes {
		topic := fmt.Sprintf("events.%s", eventType)
		messages, err := p.subscriber.Subscribe(ctx, topic)
		if err != nil {
			return fmt.Errorf("failed to subscribe to %s: %w", topic, err)
		}

		// Start handler for this topic
		go p.handleMessages(ctx, messages, eventType)
	}

	p.logger.Info("Event processor started", watermill.LogFields{
		"event_types": len(eventTypes),
	})

	// Wait for context cancellation
	<-ctx.Done()
	return nil
}

// handleMessages processes messages for a specific event type
func (p *Processor) handleMessages(ctx context.Context, messages <-chan *message.Message, eventType string) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-messages:
			if msg == nil {
				return
			}

			if err := p.processMessage(ctx, msg, eventType); err != nil {
				p.logger.Error("Failed to process message", err, watermill.LogFields{
					"event_type": eventType,
					"message_id": msg.UUID,
				})
				msg.Nack()
			} else {
				msg.Ack()
			}
		}
	}
}

// processMessage processes a single message
func (p *Processor) processMessage(ctx context.Context, msg *message.Message, eventType string) error {
	// Parse the event
	event, err := parser.ParseLine(string(msg.Payload))
	if err != nil {
		return fmt.Errorf("failed to parse event: %w", err)
	}

	if event == nil {
		// Silently skip nil events (empty lines, comments)
		return nil
	}

	// Store raw event first
	eventID, err := p.storeRawEvent(ctx, event, msg.Payload)
	if err != nil {
		return fmt.Errorf("failed to store raw event: %w", err)
	}

	// Process based on event type
	if err := p.processTypedEvent(ctx, event, eventID); err != nil {
		return fmt.Errorf("failed to process typed event: %w", err)
	}

	// Mark event as processed
	if err := p.store.MarkEventProcessed(ctx, eventID); err != nil {
		p.logger.Error("Failed to mark event as processed", err, watermill.LogFields{
			"event_id": eventID,
		})
	}

	return nil
}

// storeRawEvent stores the raw event JSON
func (p *Processor) storeRawEvent(ctx context.Context, event *events.Event, payload []byte) (int64, error) {
	var baseEvent events.BaseEvent

	// Extract base event fields
	if err := json.Unmarshal(payload, &baseEvent); err != nil {
		return 0, err
	}

	return p.store.InsertRawEvent(
		ctx,
		string(event.Type),
		baseEvent.Timestamp,
		baseEvent.ServerIP,
		baseEvent.Gamemode,
		json.RawMessage(payload),
	)
}

// processTypedEvent processes the event based on its type
func (p *Processor) processTypedEvent(ctx context.Context, event *events.Event, eventID int64) error {
	switch event.Type {
	case events.EventTypeKill:
		return p.processKillEvent(ctx, event.Kill, eventID)

	case events.EventTypeAirshot:
		return p.processAirshotEvent(ctx, event.Airshot, eventID)

	case events.EventTypeDeflect:
		return p.processDeflectEvent(ctx, event.Deflect, eventID)

	case events.EventTypeRoundStart, events.EventTypeMatchStart:
		return p.processMatchStartEvent(ctx, event.MatchStart)

	case events.EventTypeRoundEnd, events.EventTypeMatchEnd:
		return p.processMatchEndEvent(ctx, event.MatchEnd)

	default:
		// Event type stored but not processed further
		return nil
	}
}

// processKillEvent processes a kill event
func (p *Processor) processKillEvent(ctx context.Context, kill *events.KillEvent, eventID int64) error {
	// Get or create active match
	match, err := p.store.GetOrCreateActiveMatch(ctx, kill.ServerIP, "", kill.Gamemode)
	if err != nil {
		return fmt.Errorf("failed to get/create match: %w", err)
	}

	// Insert kill
	return p.store.InsertKill(ctx, kill, eventID, match.ID)
}

// processAirshotEvent processes an airshot event
func (p *Processor) processAirshotEvent(ctx context.Context, airshot *events.AirshotEvent, eventID int64) error {
	match, err := p.store.GetOrCreateActiveMatch(ctx, airshot.ServerIP, "", airshot.Gamemode)
	if err != nil {
		return err
	}

	return p.store.InsertAirshot(ctx, airshot, eventID, match.ID)
}

// processDeflectEvent processes a deflect event
func (p *Processor) processDeflectEvent(ctx context.Context, deflect *events.DeflectEvent, eventID int64) error {
	match, err := p.store.GetOrCreateActiveMatch(ctx, deflect.ServerIP, "", deflect.Gamemode)
	if err != nil {
		return err
	}

	return p.store.InsertDeflect(ctx, deflect, eventID, match.ID)
}

// processMatchStartEvent processes a match start event
func (p *Processor) processMatchStartEvent(ctx context.Context, matchStart *events.MatchStartEvent) error {
	// Create new match
	_, err := p.store.CreateMatch(ctx, matchStart.ServerIP, matchStart.Map, matchStart.Gamemode, matchStart.Timestamp)
	return err
}

// processMatchEndEvent processes a match end event
func (p *Processor) processMatchEndEvent(ctx context.Context, matchEnd *events.MatchEndEvent) error {
	// Get active match
	match, err := p.store.GetOrCreateActiveMatch(ctx, matchEnd.ServerIP, "", matchEnd.Gamemode)
	if err != nil {
		return err
	}

	// End the match
	return p.store.EndMatch(ctx, match.ID, matchEnd.WinnerTeam, 0, 0) // TODO: Get actual scores
}
