package comment

import (
	"context"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

func (s *Service) CommentByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
	comm, err := s.repo.CommentByID(ctx, id)
	if err != nil {
		// TODO: wrap in useful format
		return nil, err
	}

	return comm, nil
}
