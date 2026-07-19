package user

import (
	"context"
	"fmt"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.repo.UserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user %s: %w", id, err)
	}

	return user, nil
}

func (s *Service) GetByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domain.User, error) {
	users, err := s.repo.UsersByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("get users: %w", err)
	}

	return users, nil
}
