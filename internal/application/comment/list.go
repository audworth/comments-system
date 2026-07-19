package comment

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
	PostID   uuid.UUID
	ParentID *uuid.UUID
	Limit    int
	After    *Position
}

type Page struct {
	Comments    []domain.Comment
	EndCursor   *Position
	HasNextPage bool
}

func (s *Service) List(ctx context.Context, params *ListParams) (*Page, error) {
	if params.Limit < 1 || params.Limit > 100 {
		return nil, application.ErrInvalidPageSize
	}

	commsPage, err := s.repo.ListChildren(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list comments for post %s: %w", params.PostID, err)
	}

	return commsPage, nil
}
