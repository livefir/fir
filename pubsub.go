package fir

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"sync"

	"github.com/go-redis/redis/v8"
)

// code modeled after https://github.com/purposeinplay/go-commons/blob/v0.6.2/pubsub/inmem/pubsub.go

// Subscription is a subscription to a channel.
type Subscription interface {
	// C returns a receive-only go channel of patches published
	C() <-chan Patchset
	// Close closes the subscription.
	Close()
}

// PubsubAdapter is an interface for a pubsub adapter. It allows to publish and subscribe Patchset to views.
type PubsubAdapter interface {
	// Publish publishes a patchset to a channel.
	Publish(ctx context.Context, channel string, patchset Patchset) error
	// Subscribe subscribes to a channel.
	Subscribe(ctx context.Context, channel string) (Subscription, error)
	// HasSubscribers returns true if there are subscribers to the given pattern.
	HasSubscribers(ctx context.Context, pattern string) bool
}

// NewPubsubInmem creates a new in-memory pubsub adapter.s
func NewPubsubInmem() PubsubAdapter {
	return &pubsubInmem{
		channelsSubscriptions: make(map[string]map[*subscriptionInmem]struct{}),
	}
}

type subscriptionInmem struct {
	channel string
	ch      chan Patchset
	once    sync.Once
	pubsub  *pubsubInmem
}

// C returns a receive-only go channel of patches published
// on the channel this subscription is subscribed to.
func (s *subscriptionInmem) C() <-chan Patchset {
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
	log.Printf("removed subscribtion for channel %s, count: %d", subscription.channel, len(subscriptions))
	if len(subscriptions) == 0 {
		delete(p.channelsSubscriptions, subscription.channel)
	}
}

func (p *pubsubInmem) Publish(ctx context.Context, channel string, patchset Patchset) error {
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
		go func(sub *subscriptionInmem) { sub.ch <- patchset }(subscription)
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
		ch:      make(chan Patchset),
		pubsub:  p,
	}

	subs, ok := p.channelsSubscriptions[channel]
	if !ok {
		subs = make(map[*subscriptionInmem]struct{})
		p.channelsSubscriptions[channel] = subs
	}

	subs[sub] = struct{}{}

	log.Printf("new subscribtion for channel %s, count: %d", channel, len(subs))

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

// NewPubsubRedis creates a new redis pubsub adapter.
func NewPubsubRedis(client *redis.Client) PubsubAdapter {
	return &pubsubRedis{client: client}
}

type subscriptionRedis struct {
	channel string
	ch      chan Patchset
	once    sync.Once
	pubsub  *redis.PubSub
}

func (s *subscriptionRedis) C() <-chan Patchset {
	go func() {
		for msg := range s.pubsub.Channel() {
			var patches []patch
			err := json.Unmarshal([]byte(msg.Payload), &patches)
			if err != nil {
				log.Printf("failed to unmarshal patches payload: %v", err)
				continue
			}
			var patchset Patchset
			for _, p := range patches {
				patchset = append(patchset, p.toPatch())
			}
			s.ch <- patchset
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

func (p *pubsubRedis) Publish(ctx context.Context, channel string, patchset Patchset) error {

	var patches []patch

	for _, p := range patchset {
		patches = append(patches, patch{
			OpVal:    p.Op(),
			Selector: p.GetSelector(),
		})
	}

	patchesBytes, err := json.Marshal(patches)
	if err != nil {
		return err
	}

	return p.client.Publish(ctx, channel, patchesBytes).Err()
}

func (p *pubsubRedis) Subscribe(ctx context.Context, channel string) (Subscription, error) {
	if channel == "" {
		return nil, fmt.Errorf("channel is empty")
	}
	pubsub := p.client.Subscribe(ctx, channel)
	return &subscriptionRedis{pubsub: pubsub, channel: channel, ch: make(chan Patchset)}, nil
}

func (p *pubsubRedis) HasSubscribers(ctx context.Context, pattern string) bool {
	channels, err := p.client.PubSubChannels(ctx, pattern).Result()
	if err != nil {
		log.Printf("error getting channels for pattern: %v : err, %v", pattern, err)
		return false
	}
	if len(channels) == 0 {
		return false
	}
	return true
}