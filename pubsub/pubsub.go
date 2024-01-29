package pubsub

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/goccy/go-json"

	"github.com/go-redis/redis/v8"
	"github.com/livefir/fir/internal/eventstate"
	"github.com/livefir/fir/internal/logger"
)

// code modeled after https://github.com/purposeinplay/go-commons/blob/v0.6.2/pubsub/inmem/pubsub.go

type Event struct {
	ID          *string         `json:"id"`
	State       eventstate.Type `json:"state"`
	Target      *string         `json:"target"`
	Detail      any             `json:"detail"`
	StateDetail any             `json:"state_detail"`
	SessionID   *string         `json:"session_id"`
	ElementKey  *string         `json:"element_key"`
}

// Subscription is a subscription to a channel.
type Subscription interface {
	// C returns a receive-only go channel of events published
	C() <-chan Event
	// Close closes the subscription.
	Close()
}

// Adapter is an interface for a pubsub adapter. It allows to publish and subscribe []PubsubEvent to views.
type Adapter interface {
	// Publish publishes a events to a channel.
	Publish(ctx context.Context, channel string, event Event) error
	// Subscribe subscribes to a channel.
	Subscribe(ctx context.Context, channel string) (Subscription, error)
	// HasSubscribers returns true if there are subscribers to the given pattern.
	HasSubscribers(ctx context.Context, pattern string) bool
}

// NewInmem creates a new in-memory pubsub adapter.s
func NewInmem() Adapter {
	return &pubsubInmem{
		channelsSubscriptions: make(map[string]map[*subscriptionInmem]struct{}),
	}
}

type subscriptionInmem struct {
	channel string
	ch      chan Event
	once    sync.Once
	pubsub  *pubsubInmem
}

// C returns a receive-only go channel of events published
// on the channel this subscription is subscribed to.
func (s *subscriptionInmem) C() <-chan Event {
	return s.ch
}

func (s *subscriptionInmem) Close() {
	s.pubsub.Lock()
	defer s.pubsub.Unlock()
	s.pubsub.removeSubscription(s)
}

type pubsubInmem struct {
	channelsSubscriptions map[string]map[*subscriptionInmem]struct{}
	sync.RWMutex
}

func (p *pubsubInmem) removeSubscription(subscription *subscriptionInmem) {
	subscription.once.Do(func() {
		close(subscription.ch)
	})

	subscriptions, ok := p.channelsSubscriptions[subscription.channel]
	if !ok {
		return
	}
	delete(subscriptions, subscription)
	if len(subscriptions) == 0 {
		delete(p.channelsSubscriptions, subscription.channel)
	}
}

func (p *pubsubInmem) Publish(ctx context.Context, channel string, event Event) error {
	p.Lock()
	defer p.Unlock()
	if channel == "" {
		return fmt.Errorf("channel is empty")
	}
	subscriptions, ok := p.channelsSubscriptions[channel]
	if !ok {
		return fmt.Errorf("channel %s has no subscribers", channel)
	}
	if len(subscriptions) == 0 {
		delete(p.channelsSubscriptions, channel)
		return nil
	}

	for subscription := range subscriptions {
		go func(sub *subscriptionInmem) { sub.ch <- event }(subscription)
	}

	return nil
}

func (p *pubsubInmem) Subscribe(ctx context.Context, channel string) (Subscription, error) {
	p.Lock()
	defer p.Unlock()
	if channel == "" {
		return nil, fmt.Errorf("channel is empty")
	}

	sub := &subscriptionInmem{
		channel: channel,
		ch:      make(chan Event),
		pubsub:  p,
	}

	subs, ok := p.channelsSubscriptions[channel]
	if !ok {
		subs = make(map[*subscriptionInmem]struct{})
		p.channelsSubscriptions[channel] = subs
	}

	subs[sub] = struct{}{}

	return sub, nil
}

func (p *pubsubInmem) HasSubscribers(ctx context.Context, pattern string) bool {
	p.Lock()
	defer p.Unlock()
	count := 0
	for channel := range p.channelsSubscriptions {
		matched, err := filepath.Match(pattern, channel)
		if err != nil {
			continue
		}
		if matched {
			count++
		}
	}

	return count > 0
}

// NewRedis creates a new redis pubsub adapter.
func NewRedis(client *redis.Client) Adapter {
	return &pubsubRedis{client: client}
}

type subscriptionRedis struct {
	channel string
	ch      chan Event
	once    sync.Once
	pubsub  *redis.PubSub
}

func (s *subscriptionRedis) C() <-chan Event {
	go func() {
		for msg := range s.pubsub.Channel() {
			var events Event
			err := json.Unmarshal([]byte(msg.Payload), &events)
			if err != nil {
				logger.Errorf("failed to unmarshal events payload: %v", err)
				continue
			}
			s.ch <- events
		}
	}()
	return s.ch
}

func (s *subscriptionRedis) Close() {
	s.pubsub.Close()
	s.once.Do(func() {
		close(s.ch)
	})
}

type pubsubRedis struct {
	client *redis.Client
}

func (p *pubsubRedis) Publish(ctx context.Context, channel string, event Event) error {

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.client.Publish(ctx, channel, eventBytes).Err()
}

func (p *pubsubRedis) Subscribe(ctx context.Context, channel string) (Subscription, error) {
	if channel == "" {
		return nil, fmt.Errorf("channel is empty")
	}
	pubsub := p.client.Subscribe(ctx, channel)
	return &subscriptionRedis{pubsub: pubsub, channel: channel, ch: make(chan Event)}, nil
}

func (p *pubsubRedis) HasSubscribers(ctx context.Context, pattern string) bool {
	channels, err := p.client.PubSubChannels(ctx, pattern).Result()
	if err != nil {
		logger.Errorf("error getting channels for pattern: %v : err, %v", pattern, err)
		return false
	}
	if len(channels) == 0 {
		return false
	}
	return true
}
