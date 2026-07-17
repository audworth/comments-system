package comment

import (
	"context"
	"fmt"
	"time"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

type NewCommentParams struct {
	PostID   uuid.UUID
	ParentID *uuid.UUID
	AuthorID uuid.UUID
	Body     string
}

func (s *Service) PublishNewComment(ctx context.Context, params *NewCommentParams) (*domain.Comment, error) {
	comm, err := domain.NewComment(
		uuid.New(),
		params.PostID,
		params.ParentID,
		params.AuthorID,
		params.Body,
		time.Now().UTC(),
	)
	if err != nil {
		return nil, fmt.Errorf("invalid comment: %w", err)
	}

	created, err := s.repo.NewComment(ctx, comm)
	if err != nil {
		return nil, fmt.Errorf("publish comment: %w", err)
	}

	// TODO: handle error and log
	_ = s.notifier.NotifyCreated(ctx, created)

	return created, nil
}
