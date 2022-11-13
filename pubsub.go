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
		channelsSubs: make(map[string]map[*subscriptionInmem]struct{}),
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
	channelsSubs map[string]map[*subscriptionInmem]struct{}
	sync.RWMutex
}

func (p *pubsubInmem) removeSubscription(sub *subscriptionInmem) {
	sub.once.Do(func() {
		close(sub.ch)
	})

	subs, ok := p.channelsSubs[sub.channel]
	if !ok {
		return
	}
	delete(subs, sub)
	log.Printf("removed subscribtion for channel %s, count: %d", sub.channel, len(subs))
	if len(subs) == 0 {
		delete(p.channelsSubs, sub.channel)
	}
}

func (p *pubsubInmem) Publish(ctx context.Context, channel string, patch Patch) error {
	p.Lock()
	defer p.Unlock()
	if channel == "" {
		return fmt.Errorf("channel is empty")
	}
	subscriptions, ok := p.channelsSubs[channel]
	if !ok || len(subscriptions) == 0 {
		return fmt.Errorf("channel %s has no subscribers", channel)
	}

	for sub := range subscriptions {
		go func(sub *subscriptionInmem) { sub.ch <- patch }(sub)
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

	subs, ok := p.channelsSubs[channel]
	if !ok {
		subs = make(map[*subscriptionInmem]struct{})
		p.channelsSubs[channel] = subs
	}

	subs[sub] = struct{}{}

	log.Printf("new subscribtion for channel %s, count: %d", channel, len(subs))

	return sub, nil
}

func (p *pubsubInmem) HasSubscribers(ctx context.Context, pattern string) int {
	p.Lock()
	defer p.Unlock()
	count := 0
	for channel := range p.channelsSubs {
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
