package comment

import (
	"context"
	"fmt"
	"time"

	"github.com/audworth/comments-system/internal/application"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

type PublishParams struct {
	PostID   uuid.UUID
	ParentID *uuid.UUID
	AuthorID uuid.UUID
	Body     string
}

type Position struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

type ListParams struct {
	PostID   uuid.UUID
	ParentID *uuid.UUID
	Limit    int
	After    *Position
}

type Page struct {
	Comments    []domain.Comment
	EndCursor   *Position
	HasNextPage bool
}

//go:generate go tool mockgen -destination=mocks_test.go -package=comment . Repository,Notifier
type Repository interface {
	NewComment(ctx context.Context, comment *domain.Comment) (*domain.Comment, error)
	CommentByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error)
	ListChildren(ctx context.Context, params ListParams) (*Page, error)
	ListChildrenBatch(ctx context.Context, params []ListParams) ([]*Page, error)
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

func (s *Service) Publish(ctx context.Context, params PublishParams) (*domain.Comment, error) {
	comm, err := domain.NewComment(
		uuid.New(),
		params.PostID,
		params.ParentID,
		params.AuthorID,
		params.Body,
		time.Now().UTC(),
	)
	if err != nil {
		return nil, fmt.Errorf("invalid comment: %w", err)
	}

	created, err := s.repo.NewComment(ctx, comm)
	if err != nil {
		return nil, fmt.Errorf("publish comment: %w", err)
	}

	// TODO: handle error and log
	_ = s.notifier.NotifyCreated(ctx, created)

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
	comm, err := s.repo.CommentByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get comment %s: %w", id, err)
	}

	return comm, nil
}

func (s *Service) List(ctx context.Context, params ListParams) (*Page, error) {
	if params.Limit < 1 || params.Limit > 100 {
		return nil, application.ErrInvalidPageSize
	}

	page, err := s.repo.ListChildren(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list comments for post %s: %w", params.PostID, err)
	}

	return page, nil
}

func (s *Service) ListBatch(ctx context.Context, params []ListParams) ([]*Page, error) {
	if len(params) == 0 {
		return []*Page{}, nil
	}

	for i := range params {
		if params[i].Limit < 1 || params[i].Limit > 100 {
			return nil, fmt.Errorf(
				"invalid comment page size %d: %w",
				i,
				application.ErrInvalidPageSize,
			)
		}
	}

	pages, err := s.repo.ListChildrenBatch(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list comment pages: %w", err)
	}

	if len(pages) != len(params) {
		return nil, fmt.Errorf(
			"list comment pages: returned %d pages for %d requests",
			len(pages),
			len(params),
		)
	}

	return pages, nil
}
