package subscription

import (
	"comments-service/internal/domain"
	"context"
	"sync"
)

const subscriberBufferSize = 16

type Broker struct {
	mu          sync.Mutex
	subscribers map[int64]map[chan *domain.Comment]struct{}
}

func NewBroker() *Broker {
	return &Broker{
		subscribers: make(
			map[int64]map[chan *domain.Comment]struct{},
		),
	}
}

func (b *Broker) Subscribe(ctx context.Context, postID int64) (<-chan *domain.Comment, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	ch := make(chan *domain.Comment, subscriberBufferSize)

	b.mu.Lock()

	if _, exists := b.subscribers[postID]; !exists {
		b.subscribers[postID] = make(map[chan *domain.Comment]struct{})
	}

	b.subscribers[postID][ch] = struct{}{}

	b.mu.Unlock()

	go func() {
		<-ctx.Done()

		b.mu.Lock()
		defer b.mu.Unlock()

		subscribers := b.subscribers[postID]

		if _, exists := subscribers[ch]; !exists {
			return
		}

		delete(subscribers, ch)
		close(ch)

		if len(subscribers) == 0 {
			delete(b.subscribers, postID)
		}
	}()

	return ch, nil
}

func (b *Broker) Publish(ctx context.Context, comment *domain.Comment) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	subscribers := b.subscribers[comment.PostId]

	for ch := range subscribers {
		select {
		case ch <- comment:

		default:
			delete(subscribers, ch)
			close(ch)
		}
	}

	if len(subscribers) == 0 {
		delete(b.subscribers, comment.PostId)
	}

	return nil
}
