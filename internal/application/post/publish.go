package post

import (
	"context"
	"fmt"
	"time"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

type NewPostParams struct {
	AuthorID        uuid.UUID
	Title           string
	Body            string
	CommentsEnabled bool
}

func (s *Service) PublishNewPost(ctx context.Context, params NewPostParams) (*domain.Post, error) {
	post, err := domain.NewPost(
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

	created, err := s.repo.NewPost(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("publish post: %w", err)
	}

	return created, nil
}
