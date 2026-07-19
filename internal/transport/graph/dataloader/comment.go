package dataloader

import (
	"context"
	"time"

	"github.com/audworth/comments-system/internal/application/comment"
	"github.com/google/uuid"
	"github.com/vikstrous/dataloadgen"
)

type CommentPageKey struct {
	PostID   uuid.UUID
	ParentID uuid.UUID
	Limit    int

	AfterCreatedAt time.Time
	AfterID        uuid.UUID
}

type CommentReader interface {
	ListBatch(
		ctx context.Context,
		params []comment.ListParams,
	) ([]*comment.Page, error)
}

func newCommentPageLoader(
	comments CommentReader,
) *dataloadgen.Loader[CommentPageKey, *comment.Page] {
	return dataloadgen.NewLoader(
		func(ctx context.Context, keys []CommentPageKey) ([]*comment.Page, []error) {
			params := make([]comment.ListParams, len(keys))
			for i, key := range keys {
				params[i] = listParamsFromKey(key)
			}

			pages, err := comments.ListBatch(ctx, params)
			if err != nil {
				return nil, []error{err}
			}

			return pages, nil
		},
		dataloadgen.WithWait(loaderWait),
		dataloadgen.WithBatchCapacity(loaderBatchCapacity),
	)
}

func listParamsFromKey(key CommentPageKey) comment.ListParams {
	params := comment.ListParams{
		PostID: key.PostID,
		Limit:  key.Limit,
	}
	if key.ParentID != uuid.Nil {
		params.ParentID = &key.ParentID
	}

	if key.AfterID != uuid.Nil {
		params.After = &comment.Position{
			CreatedAt: key.AfterCreatedAt,
			ID:        key.AfterID,
		}
	}

	return params
}

func GetCommentPage(ctx context.Context, key CommentPageKey) (*comment.Page, error) {
	loaders, err := fromContext(ctx)
	if err != nil {
		return nil, err
	}

	return loaders.CommentPageLoader.Load(ctx, key)
}
