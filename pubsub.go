package fir

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
)

type PubsubAdapter interface {
	Publish(ctx context.Context, channel string, patch Patch) error
	Subscribe(ctx context.Context, channel string) (ch <-chan Patch, closeCh func() error)
	HasSubscribers(ctx context.Context, pattern string) int
}

func NewPubsubInmem() PubsubAdapter {
	return &pubsubInmem{
		channels: make(map[string]chan Patch),
	}
}

type pubsubInmem struct {
	channels map[string]chan Patch
	sync.RWMutex
}

func (p *pubsubInmem) Publish(ctx context.Context, channel string, patch Patch) error {
	p.Lock()
	defer p.Unlock()
	ch, ok := p.channels[channel]
	if !ok {
		return fmt.Errorf("channel %s not found", channel)
	}
	ch <- patch
	return nil
}

func (p *pubsubInmem) Subscribe(ctx context.Context, channel string) (<-chan Patch, func() error) {
	p.Lock()
	defer p.Unlock()
	ch := make(chan Patch)
	p.channels[channel] = ch
	return ch, func() error {
		p.Lock()
		defer p.Unlock()
		ch, ok := p.channels[channel]
		if !ok {
			return fmt.Errorf("channel %s not found", channel)
		}
		close(ch)
		delete(p.channels, channel)
		return nil
	}
}

func (p *pubsubInmem) HasSubscribers(ctx context.Context, pattern string) int {
	p.Lock()
	defer p.Unlock()
	count := 0
	for k := range p.channels {
		matched, err := filepath.Match(pattern, k)
		if err != nil {
			continue
		}
		if matched {
			count++
		}
	}

	return count
}

func NewPubsubRedis() PubsubAdapter {
	return &pubsubRedis{}
}

type pubsubRedis struct {
}

func (p *pubsubRedis) Publish(ctx context.Context, channel string, patch Patch) error {
	return nil
}

func (p *pubsubRedis) Subscribe(ctx context.Context, channel string) (ch <-chan Patch, closeCh func() error) {
	return nil, nil
}

func (p *pubsubRedis) HasSubscribers(ctx context.Context, pattern string) int {

	return 0
}
