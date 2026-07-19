package post

import (
	"context"
	"fmt"
	"time"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

type PublishParams struct {
	AuthorID        uuid.UUID
	Title           string
	Body            string
	CommentsEnabled bool
}

func (s *Service) Publish(ctx context.Context, params PublishParams) (*domain.Post, error) {
	p, err := domain.NewPost(
		uuid.New(),
		params.AuthorID,
		params.Title,
		params.Body,
		params.CommentsEnabled,
		time.Now().UTC(),
	)
	if err != nil {
		return nil, fmt.Errorf("invalid post: %w", err)
	}

	created, err := s.repo.NewPost(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("publish post: %w", err)
	}

	return created, nil
}
