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

type Subscription interface {
	C() <-chan []byte
	Close()
}

type PubsubAdapter interface {
	Publish(ctx context.Context, channel string, op Operation) error
	Subscribe(ctx context.Context, channel string) (Subscription, error)
	HasSubscribers(ctx context.Context, pattern string) bool
}

func NewPubsubInmem() PubsubAdapter {
	return &pubsubInmem{
		channelsSubscriptions: make(map[string]map[*subscriptionInmem]struct{}),
	}
}

type subscriptionInmem struct {
	channel string
	ch      chan []byte
	once    sync.Once
	pubsub  *pubsubInmem
}

// C returns a receive-only go channel of patches published
// on the channel this subscription is subscribed to.
func (s *subscriptionInmem) C() <-chan []byte {
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

func (p *pubsubInmem) Publish(ctx context.Context, channel string, op Operation) error {
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

	ops := []Operation{op}
	opsBytes, err := json.Marshal(ops)
	if err != nil {
		return err
	}

	for subscription := range subscriptions {
		go func(sub *subscriptionInmem) { sub.ch <- opsBytes }(subscription)
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
		ch:      make(chan []byte),
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

func NewPubsubRedis(client *redis.Client) PubsubAdapter {
	return &pubsubRedis{client: client}
}

type subscriptionRedis struct {
	channel string
	ch      chan []byte
	once    sync.Once
	pubsub  *redis.PubSub
}

func (s *subscriptionRedis) C() <-chan []byte {
	go func() {
		for msg := range s.pubsub.Channel() {
			s.ch <- []byte(msg.Payload)
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

func (p *pubsubRedis) Publish(ctx context.Context, channel string, op Operation) error {
	ops := []Operation{op}
	opsBytes, err := json.Marshal(ops)
	if err != nil {
		return err
	}
	return p.client.Publish(ctx, channel, opsBytes).Err()
}

func (p *pubsubRedis) Subscribe(ctx context.Context, channel string) (Subscription, error) {
	if channel == "" {
		return nil, fmt.Errorf("channel is empty")
	}
	pubsub := p.client.Subscribe(ctx, channel)
	return &subscriptionRedis{pubsub: pubsub, channel: channel, ch: make(chan []byte)}, nil
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
