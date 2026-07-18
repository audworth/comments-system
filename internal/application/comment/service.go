package comment

import (
	"context"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	NewComment(ctx context.Context, comment *domain.Comment) (*domain.Comment, error)
	CommentByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error)
	ListChildren(ctx context.Context, params *ListParams) (*Page, error)
}

type Notifier interface {
	NotifyCreated(ctx context.Context, comment *domain.Comment) error
}

type Service struct {
	repo     Repository
	notifier Notifier
}

func NewService(repo Repository, notifier Notifier) *Service {
	return &Service{
		repo:     repo,
		notifier: notifier,
	}
}
