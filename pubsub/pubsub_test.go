package pubsub

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/livefir/fir/internal/dom"
	"github.com/livefir/fir/internal/eventstate"
	"github.com/redis/go-redis/v9"
	redisContainer "github.com/testcontainers/testcontainers-go/modules/redis"
)

func ptr(s string) *string {
	return &s
}

func TestInmemPublishAndSubscribe(t *testing.T) {
	// Create a new in-memory pubsub adapter
	pubsub := NewInmem()

	// Create a channel name for testing
	channel := "test-channel"

	// Create a test event
	event := Event{
		ID:     ptr("event-id"),
		State:  eventstate.Type("event-state"),
		Target: ptr("event-target"),
		Detail: &dom.Detail{
			// Set the details of the event
		},
		SessionID:  ptr("event-session-id"),
		ElementKey: ptr("event-element-key"),
	}

	// Subscribe to the channel
	subscription, err := pubsub.Subscribe(context.Background(), channel)
	if err != nil {
		t.Fatalf("failed to subscribe to channel: %v", err)
	}

	// Publish the event to the channel
	err = pubsub.Publish(context.Background(), channel, event)
	if err != nil {
		t.Fatalf("failed to publish event: %v", err)
	}

	// Wait for the event to be received
	receivedEvent := <-subscription.C()

	// Check if the received event matches the published event
	if *receivedEvent.ID != *event.ID {
		t.Fatalf("received event ID does not match published event ID")
	}

	// Close the subscription
	subscription.Close()
}

func TestInmemHasSubscribers(t *testing.T) {
	// Create a new in-memory pubsub adapter
	pubsub := NewInmem()

	// Create a channel name for testing
	channel := "test-channel"

	// Check if the channel has subscribers
	hasSubscribers := pubsub.HasSubscribers(context.Background(), channel)
	if hasSubscribers {
		t.Fatalf("channel should not have subscribers")
	}

	// Subscribe to the channel
	subscription, err := pubsub.Subscribe(context.Background(), channel)
	if err != nil {
		t.Fatalf("failed to subscribe to channel: %v", err)
	}

	// Check if the channel has subscribers
	hasSubscribers = pubsub.HasSubscribers(context.Background(), channel)
	if !hasSubscribers {
		t.Fatalf("channel should have subscribers")
	}

	// Close the subscription
	subscription.Close()
}

func TestInmemHasSubscribersWithPattern(t *testing.T) {
	// Create a new in-memory pubsub adapter
	pubsub := NewInmem()

	// Create a channel name for testing
	channel := "test-channel"

	// Check if the channel has subscribers
	hasSubscribers := pubsub.HasSubscribers(context.Background(), channel)
	if hasSubscribers {
		t.Fatalf("channel should not have subscribers")
	}

	// Subscribe to the channel
	subscription, err := pubsub.Subscribe(context.Background(), channel)
	if err != nil {
		t.Fatalf("failed to subscribe to channel: %v", err)
	}

	// Check if the channel has subscribers
	hasSubscribers = pubsub.HasSubscribers(context.Background(), channel)
	if !hasSubscribers {
		t.Fatalf("channel should have subscribers")
	}

	// Check if the channel has subscribers with a pattern
	hasSubscribers = pubsub.HasSubscribers(context.Background(), "test-*")
	if !hasSubscribers {
		t.Fatalf("channel should have subscribers with pattern")
	}

	// Close the subscription
	subscription.Close()
}

func TestRedisPublishAndSubscribe(t *testing.T) {
	if os.Getenv("DOCKER") != "1" {
		t.Skip("Skipping testing since docker is not present")
	}

	ctx := context.Background()
	rc, err := redisContainer.Run(ctx, "redis:7")
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := rc.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	host, err := rc.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get host: %s", err)
	}

	port, err := rc.MappedPort(ctx, "6379")

	if err != nil {
		t.Fatalf("failed to get mapped port: %s", err)

	}

	addr := fmt.Sprintf("%s:%s", host, port.Port())
	if addr == "" {
		t.Fatalf("failed to get address: %s", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Create a new Redis pubsub adapter
	pubsub := NewRedis(
		client,
	)

	// Create a channel name for testing
	channel := "test-channel"

	// Create a test event
	event := Event{
		ID:     ptr("event-id"),
		State:  eventstate.Type("event-state"),
		Target: ptr("event-target"),
		Detail: &dom.Detail{
			// Set the details of the event
		},
		SessionID:  ptr("event-session-id"),
		ElementKey: ptr("event-element-key"),
	}

	// Subscribe to the channel
	subscription, err := pubsub.Subscribe(context.Background(), channel)
	if err != nil {
		t.Fatalf("failed to subscribe to channel: %v", err)
	}

	// Publish the event to the channel
	err = pubsub.Publish(context.Background(), channel, event)
	if err != nil {
		t.Fatalf("failed to publish event: %v", err)
	}

	// Wait for the event to be received
	receivedEvent := <-subscription.C()

	// Check if the received event matches the published event
	if *receivedEvent.ID != *event.ID {
		t.Fatalf("received event ID does not match published event ID %v, %v", *receivedEvent.ID, *event.ID)
	}

	// Close the subscription
	subscription.Close()
}

func TestRedisHasSubscribers(t *testing.T) {
	if os.Getenv("DOCKER") != "1" {
		t.Skip("Skipping testing since docker is not present")
	}

	ctx := context.Background()
	rc, err := redisContainer.Run(ctx, "redis:7")
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := rc.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	host, err := rc.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get host: %s", err)
	}

	port, err := rc.MappedPort(ctx, "6379")

	if err != nil {
		t.Fatalf("failed to get mapped port: %s", err)

	}

	addr := fmt.Sprintf("%s:%s", host, port.Port())
	if addr == "" {
		t.Fatalf("failed to get address: %s", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Create a new Redis pubsub adapter
	pubsub := NewRedis(
		client,
	)

	// Create a channel name for testing
	channel := "test-channel"

	// Check if the channel has subscribers
	hasSubscribers := pubsub.HasSubscribers(context.Background(), channel)
	if hasSubscribers {
		t.Fatalf("channel should not have subscribers")
	}

	// Subscribe to the channel
	subscription, err := pubsub.Subscribe(context.Background(), channel)
	if err != nil {
		t.Fatalf("failed to subscribe to channel: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Check if the channel has subscribers
	hasSubscribers = pubsub.HasSubscribers(context.Background(), channel)
	if !hasSubscribers {
		t.Fatalf("channel should have subscribers")
	}

	// Check if the channel has subscribers with a pattern
	hasSubscribers = pubsub.HasSubscribers(context.Background(), "test-*")
	if !hasSubscribers {
		t.Fatalf("channel should have subscribers with pattern")
	}

	// Close the subscription
	subscription.Close()
}
