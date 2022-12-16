package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"sync"

	"github.com/adnaan/fir/internal/dom"
	"github.com/go-redis/redis/v8"
	"github.com/golang/glog"
)

// code modeled after https://github.com/purposeinplay/go-commons/blob/v0.6.2/pubsub/inmem/pubsub.go

// Subscription is a subscription to a channel.
type Subscription interface {
	// C returns a receive-only go channel of patches published
	C() <-chan []dom.Patch
	// Close closes the subscription.
	Close()
}

// Adapter is an interface for a pubsub adapter. It allows to publish and subscribe []dom.Patch to views.
type Adapter interface {
	// Publish publishes a patchset to a channel.
	Publish(ctx context.Context, channel string, patchset ...dom.Patch) error
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
	ch      chan []dom.Patch
	once    sync.Once
	pubsub  *pubsubInmem
}

// C returns a receive-only go channel of patches published
// on the channel this subscription is subscribed to.
func (s *subscriptionInmem) C() <-chan []dom.Patch {
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

func (p *pubsubInmem) Publish(ctx context.Context, channel string, patchset ...dom.Patch) error {
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
		ch:      make(chan []dom.Patch),
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

// NewRedis creates a new redis pubsub adapter.
func NewRedis(client *redis.Client) Adapter {
	return &pubsubRedis{client: client}
}

type subscriptionRedis struct {
	channel string
	ch      chan []dom.Patch
	once    sync.Once
	pubsub  *redis.PubSub
}

func (s *subscriptionRedis) C() <-chan []dom.Patch {
	go func() {
		for msg := range s.pubsub.Channel() {
			var patchset []dom.Patch
			err := json.Unmarshal([]byte(msg.Payload), &patchset)
			if err != nil {
				glog.Errorf("failed to unmarshal patches payload: %v", err)
				continue
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

func (p *pubsubRedis) Publish(ctx context.Context, channel string, patchset ...dom.Patch) error {

	patchesBytes, err := json.Marshal(patchset)
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
	return &subscriptionRedis{pubsub: pubsub, channel: channel, ch: make(chan []dom.Patch)}, nil
}

func (p *pubsubRedis) HasSubscribers(ctx context.Context, pattern string) bool {
	channels, err := p.client.PubSubChannels(ctx, pattern).Result()
	if err != nil {
		glog.Errorf("error getting channels for pattern: %v : err, %v", pattern, err)
		return false
	}
	if len(channels) == 0 {
		return false
	}
	return true
}
