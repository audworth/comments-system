package user

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

//go:generate go tool mockgen -destination=mocks_test.go -package=user . Repository
type Repository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domain.User, error)
}

type Service struct {
	logger *slog.Logger
	repo   Repository
}

func NewService(repo Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	s.logger.DebugContext(ctx, "get user", slog.String("user_id", id.String()))

	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user %s: %w", id, err)
	}

	return u, nil
}

func (s *Service) GetByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domain.User, error) {
	s.logger.DebugContext(ctx, "get users", slog.Int("user_count", len(ids)))

	users, err := s.repo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("get users: %w", err)
	}

	return users, nil
}
