package post

//go:generate go tool mockgen -destination=mocks_test.go -package=post . Repository

import (
	"context"
	"fmt"
	"time"

	"github.com/audworth/comments-system/internal/application"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

type PublishParams struct {
	AuthorID        uuid.UUID
	Title           string
	Body            string
	CommentsEnabled bool
}

type Position struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

type ListParams struct {
	Limit int
	After *Position
}

type Page struct {
	Posts       []domain.Post
	Next        *Position
	HasNextPage bool
}

type Repository interface {
	Publish(ctx context.Context, post *domain.Post) (*domain.Post, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error)
	List(ctx context.Context, params ListParams) (*Page, error)
	SetCommentsEnabled(ctx context.Context, postID uuid.UUID, authorID uuid.UUID, enabled bool) (*domain.Post, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Publish(ctx context.Context, params PublishParams) (*domain.Post, error) {
	p, err := domain.NewPost(
		uuid.New(),
		params.AuthorID,
		params.Title,
		params.Body,
		params.CommentsEnabled,
		time.Now().UTC(),
	)
	if err != nil {
		return nil, fmt.Errorf("invalid post: %w", err)
	}

	created, err := s.repo.Publish(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("publish post: %w", err)
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get post %s: %w", id, err)
	}

	return p, nil
}

func (s *Service) List(ctx context.Context, params ListParams) (*Page, error) {
	if params.Limit < 1 || params.Limit > 100 {
		return nil, application.ErrInvalidPageSize
	}

	page, err := s.repo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}

	return page, nil
}

func (s *Service) SetCommentsEnabled(
	ctx context.Context,
	postID uuid.UUID,
	authorID uuid.UUID,
	enabled bool,
) (*domain.Post, error) {
	p, err := s.repo.SetCommentsEnabled(ctx, postID, authorID, enabled)
	if err != nil {
		return nil, fmt.Errorf(
			"set comments enabled for post %s (author %s): %w",
			postID,
			authorID,
			err,
		)
	}

	return p, nil
}
