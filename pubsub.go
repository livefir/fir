package fir

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
)

type PubsubAdapter interface {
	Publish(ctx context.Context, channel string, patchset Patchset) error
	Subscribe(ctx context.Context, channel string) (ch <-chan Patchset, closeCh func() error)
	HasSubscribers(ctx context.Context, pattern string) int
}

func NewPubsubInmem() PubsubAdapter {
	return &pubsubInmem{
		channels: make(map[string]chan Patchset),
	}
}

type pubsubInmem struct {
	channels map[string]chan Patchset
	sync.RWMutex
}

func (p *pubsubInmem) Publish(ctx context.Context, channel string, patchset Patchset) error {
	p.Lock()
	defer p.Unlock()
	ch, ok := p.channels[channel]
	if !ok {
		return fmt.Errorf("channel %s not found", channel)
	}
	ch <- patchset
	return nil
}

func (p *pubsubInmem) Subscribe(ctx context.Context, channel string) (<-chan Patchset, func() error) {
	p.Lock()
	defer p.Unlock()
	ch := make(chan Patchset)
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

func (p *pubsubRedis) Publish(ctx context.Context, channel string, patchset Patchset) error {
	return nil
}

func (p *pubsubRedis) Subscribe(ctx context.Context, channel string) (ch <-chan Patchset, closeCh func() error) {
	return nil, nil
}

func (p *pubsubRedis) HasSubscribers(ctx context.Context, pattern string) int {

	return 0
}
