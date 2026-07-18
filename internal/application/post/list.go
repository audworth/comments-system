package post

import (
	"context"
	"fmt"
	"time"

	"github.com/audworth/comments-system/internal/application"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

type Position struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

type ListParams struct {
	Limit int
	After *Position
}

type Page struct {
	Posts       []domain.Post
	Next        *Position
	HasNextPage bool
}

func (s *Service) ListPosts(ctx context.Context, params ListParams) (*Page, error) {
	if params.Limit < 1 || params.Limit > 100 {
		return nil, application.ErrInvalidPageSize
	}

	postPage, err := s.repo.ListPosts(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}

	return postPage, nil
}
