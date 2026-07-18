package resolver

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

import (
	"context"

	postapp "github.com/audworth/comments-system/internal/application/post"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

// PostService describes every post operation exposed by GraphQL. Depending on
// an interface keeps the transport independently unit-testable.
type PostService interface {
	PublishNewPost(ctx context.Context, params postapp.NewPostParams) (*domain.Post, error)
	PostByID(ctx context.Context, id uuid.UUID) (*domain.Post, error)
	List(ctx context.Context, params postapp.ListParams) (*postapp.Page, error)
	SetCommentsToEnabled(ctx context.Context, postID uuid.UUID, authorID uuid.UUID, enabled bool) (*domain.Post, error)
}

type Resolver struct {
	posts PostService
}

func New(posts PostService) *Resolver {
	return &Resolver{posts: posts}
}
