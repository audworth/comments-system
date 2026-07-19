package dataloader

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/audworth/comments-system/internal/application/comment"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	"github.com/vikstrous/dataloadgen"
)

const (
	loaderWait          = 2 * time.Millisecond
	loaderBatchCapacity = 100
)

var ErrNotInContext = errors.New("dataloaders are missing")

type Loaders struct {
	UserLoader        *dataloadgen.Loader[uuid.UUID, *domain.User]
	CommentPageLoader *dataloadgen.Loader[CommentPageKey, *comment.Page]
}

func NewLoaders(users userReader, comments commentReader) *Loaders {
	return &Loaders{
		UserLoader:        newUserLoader(users),
		CommentPageLoader: newCommentPageLoader(comments),
	}
}

func Middleware(users userReader, comments commentReader, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), loadersKey, NewLoaders(users, comments))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type ctxKey string

const loadersKey = ctxKey("loaders")

func fromContext(ctx context.Context) (*Loaders, error) {
	loaders, ok := ctx.Value(loadersKey).(*Loaders)
	if !ok || loaders == nil {
		return nil, ErrNotInContext
	}

	return loaders, nil
}
