package mem

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/audworth/comments-system/internal/application/comment"
	"github.com/audworth/comments-system/internal/application/user"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/audworth/comments-system/internal/platform/db/inmem"
	"github.com/google/uuid"
)

var _ comment.Repository = (*CommentsRepository)(nil)

type CommentsRepository struct {
	db     *inmem.InMem
	logger *slog.Logger
}

func NewCommentsRepository(db *inmem.InMem, logger *slog.Logger) *CommentsRepository {
	return &CommentsRepository{
		db:     db,
		logger: logger,
	}
}

func (r *CommentsRepository) Publish(_ context.Context, newComment *domain.Comment) (*domain.Comment, error) {
	r.logger.Debug(
		"publish comment",
		slog.String("comment_id", newComment.ID.String()),
		slog.String("post_id", newComment.PostID.String()),
		slog.String("author_id", newComment.AuthorID.String()),
		slog.Bool("has_parent", newComment.ParentID != nil),
	)

	comm := *newComment

	err := r.db.Update(func(tx *inmem.Tx) error {
		p, found := tx.Post(comm.PostID)
		if !found {
			return comment.ErrPostNotFound
		}
		if !p.CommentsEnabled {
			return comment.ErrCommentsDisabled
		}
		if comm.ParentID != nil {
			parent, found := tx.Comment(*comm.ParentID)
			if !found || parent.PostID != comm.PostID {
				return comment.ErrParentNotFound
			}
		}
		if _, found := tx.User(comm.AuthorID); !found {
			return user.ErrNotFound
		}
		if _, found := tx.Comment(comm.ID); found {
			return fmt.Errorf("comment %s already exists", comm.ID)
		}

		tx.PutComment(comm)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &comm, nil
}

func (r *CommentsRepository) GetByID(_ context.Context, id uuid.UUID) (*domain.Comment, error) {
	r.logger.Debug("get comment", slog.String("comment_id", id.String()))

	var (
		comm  domain.Comment
		found bool
	)
	_ = r.db.View(func(tx *inmem.Tx) error {
		comm, found = tx.Comment(id)
		return nil
	})

	if !found {
		return nil, comment.ErrNotFound
	}
	return &comm, nil
}

func (r *CommentsRepository) List(_ context.Context, params comment.ListParams) (*comment.Page, error) {
	r.logger.Debug(
		"list comments",
		slog.String("post_id", params.PostID.String()),
		slog.Bool("has_parent", params.ParentID != nil),
		slog.Int("limit", params.Limit),
		slog.Bool("has_after", params.After != nil),
	)

	var page *comment.Page
	_ = r.db.View(func(tx *inmem.Tx) error {
		page = makeCommentPage(tx, params)
		return nil
	})

	return page, nil
}

func (r *CommentsRepository) ListBatch(_ context.Context, params []comment.ListParams) ([]*comment.Page, error) {
	r.logger.Debug("list comment pages", slog.Int("batch_size", len(params)))

	pages := make([]*comment.Page, len(params))
	_ = r.db.View(func(tx *inmem.Tx) error {
		for i := range params {
			pages[i] = makeCommentPage(tx, params[i])
		}
		return nil
	})

	return pages, nil
}

func makeCommentPage(tx *inmem.Tx, params comment.ListParams) *comment.Page {
	ids := tx.CommentIDs(params.PostID, params.ParentID)
	start := 0
	if params.After != nil {
		start = sort.Search(len(ids), func(i int) bool {
			comm, _ := tx.Comment(ids[i])
			return positionBefore(
				comm.CreatedAt,
				comm.ID,
				params.After.CreatedAt,
				params.After.ID,
			)
		})
	}

	end := min(start+params.Limit+1, len(ids))
	comments := make([]domain.Comment, 0, end-start)
	for _, id := range ids[start:end] {
		comm, _ := tx.Comment(id)
		comments = append(comments, comm)
	}

	hasNextPage := len(comments) > params.Limit
	if hasNextPage {
		comments = comments[:params.Limit]
	}

	page := &comment.Page{
		Comments:    comments,
		HasNextPage: hasNextPage,
	}
	if len(comments) > 0 {
		last := comments[len(comments)-1]
		page.EndCursor = &comment.Position{
			CreatedAt: last.CreatedAt,
			ID:        last.ID,
		}
	}

	return page
}
