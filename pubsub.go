package fir

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"sync"
)

// code modeled after https://github.com/purposeinplay/go-commons/blob/v0.6.2/pubsub/inmem/pubsub.go

type Subscription interface {
	C() <-chan Patch
	Close()
}

type PubsubAdapter interface {
	Publish(ctx context.Context, channel string, patch Patch) error
	Subscribe(ctx context.Context, channel string) (Subscription, error)
	HasSubscribers(ctx context.Context, pattern string) int
}

func NewPubsubInmem() PubsubAdapter {
	return &pubsubInmem{
		channelsSubscriptions: make(map[string]map[*subscriptionInmem]struct{}),
	}
}

type subscriptionInmem struct {
	channel string
	ch      chan Patch
	once    sync.Once
	pubsub  *pubsubInmem
}

// C returns a receive-only go channel of patches published
// on the channel this subscription is subscribed to.
func (s *subscriptionInmem) C() <-chan Patch {
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

func (p *pubsubInmem) Publish(ctx context.Context, channel string, patch Patch) error {
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
		go func(sub *subscriptionInmem) { sub.ch <- patch }(subscription)
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
		ch:      make(chan Patch),
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

func (p *pubsubInmem) HasSubscribers(ctx context.Context, pattern string) int {
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

	return count
}

type subscriptionRedis struct {
}

type pubsubRedis struct {
}
