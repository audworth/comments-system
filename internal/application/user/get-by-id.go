package user

import (
	"context"
	"fmt"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

func (s *Service) UserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.repo.UserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user %s: %w", id, err)
	}

	return user, nil
}
