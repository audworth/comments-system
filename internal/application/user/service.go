package user

import (
	"context"
	"fmt"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

//go:generate go tool mockgen -destination=mocks_test.go -package=user . Repository
type Repository interface {
	UserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UsersByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domain.User, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, err := s.repo.UserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user %s: %w", id, err)
	}

	return u, nil
}

func (s *Service) GetByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domain.User, error) {
	users, err := s.repo.UsersByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("get users: %w", err)
	}

	return users, nil
}
