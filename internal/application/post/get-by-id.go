package post

import (
	"context"
	"fmt"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

func (s *Service) PostByID(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
	post, err := s.repo.PostByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get post %s: %w", id, err)
	}

	return post, err
}
