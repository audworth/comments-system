package user

import (
	"context"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

//go:generate go tool mockgen -destination=mocks_test.go -package=user . Repository
type Repository interface {
	UserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}
