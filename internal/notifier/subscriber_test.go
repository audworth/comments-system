package notifier

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestSubscriberForwardsEventToPostSubscribers(t *testing.T) {
	subscriber := newTestSubscriber()
	defer subscriber.closeSubscribers()

	postID := uuid.New()
	otherPostID := uuid.New()
	first, err := subscriber.SubscribeCommentCreated(t.Context(), postID)
	require.NoError(t, err)
	second, err := subscriber.SubscribeCommentCreated(t.Context(), postID)
	require.NoError(t, err)
	other, err := subscriber.SubscribeCommentCreated(t.Context(), otherPostID)
	require.NoError(t, err)

	created := publishTestEvent(t, subscriber, postID)

	require.Equal(t, created, <-first)
	require.Equal(t, created, <-second)
	select {
	case <-time.After(10 * time.Second):
		t.Fatalf("unexpected deadlock")
	case unexpected := <-other:
		t.Fatalf("subscriber received event for another post: %s", unexpected.PostID)
	default:
	}
}

func TestSubscriberClosesLocalSubscriptionOnContextCancellation(t *testing.T) {
	subscriber := newTestSubscriber()
	defer subscriber.closeSubscribers()

	ctx, cancel := context.WithCancel(t.Context())
	comments, err := subscriber.SubscribeCommentCreated(ctx, uuid.New())
	require.NoError(t, err)
	cancel()

	select {
	case _, open := <-comments:
		require.False(t, open)
	case <-time.After(time.Second):
		t.Fatal("local subscription was not closed")
	}
}

func newTestSubscriber() *Subscriber {
	return &Subscriber{
		logger: slog.New(slog.DiscardHandler),
		done:   make(chan struct{}),
		subs:   make(map[uuid.UUID]map[chan *domain.Comment]struct{}),
	}
}

func publishTestEvent(t *testing.T, subscriber *Subscriber, postID uuid.UUID) *domain.Comment {
	t.Helper()

	event := commentCreatedEvent{
		ID:        uuid.New(),
		PostID:    postID,
		AuthorID:  uuid.New(),
		Body:      "body",
		CreatedAt: time.Now().UTC(),
	}
	payload, err := json.Marshal(event)
	require.NoError(t, err)

	subscriber.forward(t.Context(), &goredis.Message{
		Channel: commentCreatedTopic,
		Payload: string(payload),
	})

	created, err := domain.NewComment(
		event.ID,
		event.PostID,
		event.ParentID,
		event.AuthorID,
		event.Body,
		event.CreatedAt,
	)
	require.NoError(t, err)
	return created
}
