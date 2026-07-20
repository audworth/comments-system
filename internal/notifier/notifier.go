package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/audworth/comments-system/internal/application/comment"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

var _ comment.Notifier = (*Notifier)(nil)

type Notifier struct {
	client *goredis.Client
	logger *slog.Logger
}

func NewNotifier(client *goredis.Client, logger *slog.Logger) *Notifier {
	return &Notifier{
		client: client,
		logger: logger,
	}
}

const commentCreatedTopicPrefix = "comment.created.post."

type CommentCreatedEvent struct {
	ID        uuid.UUID  `json:"id"`
	PostID    uuid.UUID  `json:"postId"`
	ParentID  *uuid.UUID `json:"parentId,omitempty"`
	AuthorID  uuid.UUID  `json:"authorId"`
	Body      string     `json:"body"`
	CreatedAt time.Time  `json:"created_at"`
}

func (n *Notifier) NotifyCommentCreated(ctx context.Context, created *domain.Comment) error {
	e := &CommentCreatedEvent{
		ID:        created.ID,
		PostID:    created.PostID,
		ParentID:  created.ParentID,
		AuthorID:  created.AuthorID,
		Body:      created.Body,
		CreatedAt: created.CreatedAt,
	}

	jsoned, err := json.Marshal(e)
	if err != nil {
		n.logger.ErrorContext(
			ctx,
			"failed to marshal comment notification",
			slog.String("comment_id", created.ID.String()),
			slog.String("post_id", created.PostID.String()),
			slog.Any("error", err),
		)
		return fmt.Errorf("marshal created comment event: %w", err)
	}

	topic := commentCreatedTopicPrefix + created.PostID.String()
	if err := n.client.Publish(ctx, topic, jsoned).Err(); err != nil {
		n.logger.ErrorContext(
			ctx,
			"failed to publish comment notification",
			slog.String("topic", topic),
			slog.String("comment_id", created.ID.String()),
			slog.String("post_id", created.PostID.String()),
			slog.Any("error", err),
		)
		return fmt.Errorf("notify comment created: %w", err)
	}

	return nil
}
