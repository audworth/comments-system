package resolver

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

import (
	"context"

	"github.com/audworth/comments-system/internal/application/comment"
	"github.com/audworth/comments-system/internal/application/post"
	"github.com/audworth/comments-system/internal/application/user"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

//go:generate go tool mockgen -destination=mocks_test.go -package=resolver . PostsService,UsersService,CommentsService
type PostsService interface {
	Publish(ctx context.Context, params post.PublishParams) (*domain.Post, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error)
	List(ctx context.Context, params post.ListParams) (*post.Page, error)
	SetCommentsEnabled(ctx context.Context, postID uuid.UUID, authorID uuid.UUID, enabled bool) (*domain.Post, error)
}

type UsersService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type CommentsService interface {
	Publish(ctx context.Context, params comment.PublishParams) (*domain.Comment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error)
	List(ctx context.Context, params comment.ListParams) (*comment.Page, error)
	SubscribeToPostComments(ctx context.Context, postID uuid.UUID) (<-chan *domain.Comment, error)
}

var (
	_ PostsService    = (*post.Service)(nil)
	_ UsersService    = (*user.Service)(nil)
	_ CommentsService = (*comment.Service)(nil)
)

type Resolver struct {
	posts    PostsService
	users    UsersService
	comments CommentsService
}

func New(posts PostsService, users UsersService, comments CommentsService) *Resolver {
	return &Resolver{
		posts:    posts,
		users:    users,
		comments: comments,
	}
}
