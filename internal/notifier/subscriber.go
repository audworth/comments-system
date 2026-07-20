package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/audworth/comments-system/internal/application/comment"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

var _ comment.Subscriber = (*Subscriber)(nil)

type Subscriber struct {
	client *goredis.Client
	logger *slog.Logger
}

func NewSubscriber(client *goredis.Client, logger *slog.Logger) *Subscriber {
	return &Subscriber{
		client: client,
		logger: logger,
	}
}

func (s *Subscriber) SubscribeCommentCreated(ctx context.Context, postID uuid.UUID) (<-chan *domain.Comment, error) {
	topic := commentCreatedTopicPrefix + postID.String()
	s.logger.DebugContext(
		ctx,
		"subscribe to comment notifications",
		slog.String("topic", topic),
		slog.String("post_id", postID.String()),
	)

	ps := s.client.Subscribe(ctx, topic)
	if _, err := ps.Receive(ctx); err != nil {
		_ = ps.Close()
		s.logger.ErrorContext(
			ctx,
			"could not establish subscription to redis",
			slog.String("topic", topic),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("subscribe to topic %s: %w", topic, err)
	}
	s.logger.DebugContext(
		ctx,
		"comment notification subscription opened",
		slog.String("topic", topic),
		slog.String("post_id", postID.String()),
	)

	out := make(chan *domain.Comment)
	msgs := ps.Channel()

	go s.forward(ctx, ps, msgs, out, postID)

	return out, nil
}

func (s *Subscriber) forward(
	ctx context.Context,
	ps *goredis.PubSub,
	msgs <-chan *goredis.Message,
	comms chan<- *domain.Comment,
	postID uuid.UUID,
) {
	closeReason := "source_closed"
	defer func() {
		close(comms)
		if err := ps.Close(); err != nil {
			s.logger.WarnContext(
				ctx,
				"failed to close pubsub for post",
				slog.String("post_id", postID.String()),
				slog.Any("error", err),
			)
		}
		s.logger.DebugContext(
			ctx,
			"comment notification subscription closed",
			slog.String("post_id", postID.String()),
			slog.String("reason", closeReason),
		)
	}()

	for {
		select {
		case <-ctx.Done():
			closeReason = "context_done"
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}

			var e CommentCreatedEvent
			if err := json.Unmarshal([]byte(msg.Payload), &e); err != nil {
				s.logger.ErrorContext(
					ctx,
					"failed to decode event",
					slog.String("topic", msg.Channel),
					slog.Any("error", err),
				)
				continue
			}
			if e.PostID != postID {
				s.logger.WarnContext(
					ctx,
					"unexpected post_id",
					slog.String("want", postID.String()),
					slog.String("got", e.PostID.String()),
				)
				continue
			}
			s.logger.DebugContext(
				ctx,
				"comment notification received",
				slog.String("topic", msg.Channel),
				slog.String("comment_id", e.ID.String()),
				slog.String("post_id", e.PostID.String()),
			)

			comm, err := domain.NewComment(
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
				continue
			}

			select {
			case <-ctx.Done():
				closeReason = "context_done"
				return
			case comms <- comm:
				s.logger.DebugContext(
					ctx,
					"comment notification forwarded",
					slog.String("comment_id", comm.ID.String()),
					slog.String("post_id", comm.PostID.String()),
				)
			}
		}
	}
}
