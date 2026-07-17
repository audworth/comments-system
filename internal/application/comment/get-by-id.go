package comment

import (
	"context"
	"fmt"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

func (s *Service) CommentByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
	comm, err := s.repo.CommentByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get comment %s: %w", id, err)
	}

	return comm, nil
}
