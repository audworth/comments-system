package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/audworth/comments-system/internal/application/comment"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

var _ comment.Subscriber = (*Subscriber)(nil)

type Subscriber struct {
	pubsub *goredis.PubSub
	logger *slog.Logger
	cancel context.CancelFunc
	done   chan struct{}

	mu     sync.Mutex
	subs   map[uuid.UUID]map[chan *domain.Comment]struct{}
	closed bool
}

func NewSubscriber(ctx context.Context, client *goredis.Client, logger *slog.Logger) (*Subscriber, error) {
	pubsub := client.Subscribe(ctx, commentCreatedTopic)
	if _, err := pubsub.Receive(ctx); err != nil {
		_ = pubsub.Close()
		return nil, fmt.Errorf("subscribe to topic %s: %w", commentCreatedTopic, err)
	}

	subCtx, cancel := context.WithCancel(ctx)
	s := &Subscriber{
		pubsub: pubsub,
		logger: logger,
		cancel: cancel,
		done:   make(chan struct{}),
		subs:   make(map[uuid.UUID]map[chan *domain.Comment]struct{}),
	}

	go s.receive(subCtx)

	return s, nil
}

func (s *Subscriber) SubscribeCommentCreated(ctx context.Context, postID uuid.UUID) (<-chan *domain.Comment, error) {
	out := make(chan *domain.Comment, 1)

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, fmt.Errorf("comment subscriber is closed")
	}

	if s.subs[postID] == nil {
		s.subs[postID] = make(map[chan *domain.Comment]struct{})
	}
	s.subs[postID][out] = struct{}{}
	s.mu.Unlock()

	go func() {
		select {
		case <-ctx.Done():
			s.remove(postID, out)
		case <-s.done:
		}
	}()

	return out, nil
}

func (s *Subscriber) Close() {
	s.cancel()
	<-s.done
}

func (s *Subscriber) receive(ctx context.Context) {
	defer func() {
		s.closeSubscribers()
		_ = s.pubsub.Close()
	}()

	messages := s.pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-messages:
			if !ok {
				return
			}
			s.forward(ctx, msg)
		}
	}
}

func (s *Subscriber) forward(ctx context.Context, msg *goredis.Message) {
	var e commentCreatedEvent
	if err := json.Unmarshal([]byte(msg.Payload), &e); err != nil {
		s.logger.ErrorContext(
			ctx,
			"failed to decode comment event",
			slog.String("topic", msg.Channel),
			slog.Any("error", err),
		)
		return
	}

	created, err := domain.NewComment(
		e.ID,
		e.PostID,
		e.ParentID,
		e.AuthorID,
		e.Body,
		e.CreatedAt,
	)
	if err != nil {
		s.logger.ErrorContext(
			ctx,
			"invalid comment in pubsub",
			slog.String("topic", msg.Channel),
			slog.Any("error", err),
		)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for sub := range s.subs[e.PostID] {
		select {
		case sub <- created:
		default:
		}
	}
}

func (s *Subscriber) remove(postID uuid.UUID, subscriber chan *domain.Comment) {
	s.mu.Lock()
	defer s.mu.Unlock()

	postSubs := s.subs[postID]
	if _, ok := postSubs[subscriber]; !ok {
		return
	}

	delete(postSubs, subscriber)
	close(subscriber)
	if len(postSubs) == 0 {
		delete(s.subs, postID)
	}
}

func (s *Subscriber) closeSubscribers() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closed = true
	for _, postSubscribers := range s.subs {
		for subscriber := range postSubscribers {
			close(subscriber)
		}
	}
	s.subs = nil
	close(s.done)
}
