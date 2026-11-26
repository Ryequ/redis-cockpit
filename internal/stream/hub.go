package stream

import (
	"context"
	"sync"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Message is one Pub/Sub message.
type Message struct {
	Channel    string
	Payload    string
	ReceivedAt time.Time
}

// Subscription wraps a go-redis PubSub subscription with a bounded queue.
type Subscription struct {
	pubsub   *goredis.PubSub
	messages *Bounded[Message]
	cancel   context.CancelFunc
	done     chan struct{}
	once     sync.Once
}

func (s *Subscription) Messages() <-chan Message {
	return s.messages.C()
}

func (s *Subscription) Dropped() int {
	return s.messages.Dropped()
}

func (s *Subscription) Done() <-chan struct{} {
	return s.done
}

func (s *Subscription) Close() error {
	if s == nil {
		return nil
	}
	var err error
	s.once.Do(func() {
		close(s.done)
		if s.cancel != nil {
			s.cancel()
		}
		if s.pubsub != nil {
			err = s.pubsub.Close()
		}
	})
	return err
}

// Hub provides Pub/Sub and MONITOR capabilities.
type Hub struct {
	client goredis.UniversalClient
}

func NewHub(client goredis.UniversalClient) *Hub {
	return &Hub{client: client}
}

func (h *Hub) Subscribe(ctx context.Context, channels ...string) (*Subscription, error) {
	if h == nil || h.client == nil {
		return nil, errNoClient
	}
	pubsub := h.client.Subscribe(context.Background(), channels...)
	handshakeCtx, handshakeCancel := context.WithTimeout(ctx, 10*time.Second)
	defer handshakeCancel()
	if _, err := pubsub.Receive(handshakeCtx); err != nil {
		_ = pubsub.Close()
		return nil, err
	}

	subCtx, cancel := context.WithCancel(context.Background())
	subscription := &Subscription{
		pubsub:   pubsub,
		messages: NewBounded[Message](200),
		cancel:   cancel,
		done:     make(chan struct{}),
	}

	go func() {
		for {
			message, err := pubsub.ReceiveMessage(subCtx)
			if err != nil {
				return
			}
			subscription.messages.Send(subCtx, Message{
				Channel:    message.Channel,
				Payload:    message.Payload,
				ReceivedAt: time.Now(),
			})
		}
	}()
	return subscription, nil
}

func (h *Hub) Publish(ctx context.Context, channel, payload string) error {
	if h == nil || h.client == nil {
		return errNoClient
	}
	return h.client.Publish(ctx, channel, payload).Err()
}
