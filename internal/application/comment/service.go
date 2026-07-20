package comment

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/audworth/comments-system/internal/application"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

const notifyTimeout = 100 * time.Second

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
	Publish(ctx context.Context, comment *domain.Comment) (*domain.Comment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error)
	List(ctx context.Context, params ListParams) (*Page, error)
	ListBatch(ctx context.Context, params []ListParams) ([]*Page, error)
}

type Notifier interface {
	NotifyCommentCreated(ctx context.Context, comment *domain.Comment) error
}

type Service struct {
	logger   *slog.Logger
	repo     Repository
	notifier Notifier
}

func NewService(repo Repository, notifier Notifier, logger *slog.Logger) *Service {
	return &Service{
		repo:     repo,
		notifier: notifier,
		logger:   logger,
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

	created, err := s.repo.Publish(ctx, comm)
	if err != nil {
		return nil, fmt.Errorf("publish comment: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, notifyTimeout)
	defer cancel()
	if err := s.notifier.NotifyCommentCreated(ctx, created); err != nil {
		s.logger.ErrorContext(
			ctx,
			"failed to notify about new published comment",
			slog.Any("error", err),
		)
	}

	s.logger.InfoContext(
		ctx,
		"published new comment",
		slog.String("author_id", created.AuthorID.String()),
		slog.String("comment_id", created.ID.String()),
		slog.String("post_id", created.PostID.String()),
	)
	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
	comm, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get comment %s: %w", id, err)
	}

	s.logger.InfoContext(
		ctx,
		"retrieved comment",
		slog.String("author_id", comm.AuthorID.String()),
		slog.String("comment_id", comm.ID.String()),
		slog.String("post_id", comm.PostID.String()),
	)
	return comm, nil
}

func (s *Service) List(ctx context.Context, params ListParams) (*Page, error) {
	if params.Limit < 1 || params.Limit > 100 {
		return nil, application.ErrInvalidPageSize
	}

	page, err := s.repo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list comments for post %s: %w", params.PostID, err)
	}

	s.logger.InfoContext(
		ctx,
		"retrieved page of comments",
		slog.Int("amount of comments", len(page.Comments)),
	)
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

	pages, err := s.repo.ListBatch(ctx, params)
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

	s.logger.InfoContext(
		ctx,
		"retrieved batch of pages of comments",
		slog.Int("amount of pages", len(pages)),
	)
	return pages, nil
}
