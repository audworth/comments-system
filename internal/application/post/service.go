package post

import (
	"context"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	NewPost(ctx context.Context, post *domain.Post) (*domain.Post, error)
	PostByID(ctx context.Context, id uuid.UUID) (*domain.Post, error)
	List(ctx context.Context, params ListParams) (*Page, error)
	SetCommentsEnabled(ctx context.Context, postID uuid.UUID, author uuid.UUID, enabled bool) (*domain.Post, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}
