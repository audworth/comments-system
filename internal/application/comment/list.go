package comment

import (
	"context"
	"time"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

type CommentPosition struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

type ListParams struct {
	PostID   uuid.UUID
	ParentID *uuid.UUID
	Limit    int
	After    *CommentPosition
}

type Page struct {
	Comments  []domain.Comment
	EndCursor *CommentPosition
	HasNext   bool
}

func (s *Service) List(
	ctx context.Context,
	params ListParams,
) (*Page, error) {
	commsPage, err := s.repo.ListChildren(ctx, params)
	if err != nil {
		// TODO: wrap in useful format
		return nil, err
	}

	return commsPage, nil
}
