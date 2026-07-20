package mem

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/audworth/comments-system/internal/application/post"
	"github.com/audworth/comments-system/internal/application/user"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/audworth/comments-system/internal/platform/db/inmem"
	"github.com/google/uuid"
)

var _ post.Repository = (*PostRepository)(nil)

type PostRepository struct {
	db     *inmem.InMem
	logger *slog.Logger
}

func NewPostRepository(db *inmem.InMem, logger *slog.Logger) *PostRepository {
	return &PostRepository{
		db:     db,
		logger: logger,
	}
}

func (r *PostRepository) Publish(_ context.Context, newPost *domain.Post) (*domain.Post, error) {
	r.logger.Debug(
		"publish post",
		slog.String("post_id", newPost.ID.String()),
		slog.String("author_id", newPost.AuthorID.String()),
	)

	p := *newPost

	err := r.db.Update(func(tx *inmem.Tx) error {
		if _, found := tx.User(p.AuthorID); !found {
			return user.ErrNotFound
		}
		if _, found := tx.Post(p.ID); found {
			return fmt.Errorf("post %s already exists", p.ID)
		}

		tx.PutPost(p)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *PostRepository) GetByID(_ context.Context, id uuid.UUID) (*domain.Post, error) {
	r.logger.Debug("get post", slog.String("post_id", id.String()))

	var (
		p     domain.Post
		found bool
	)
	_ = r.db.View(func(tx *inmem.Tx) error {
		p, found = tx.Post(id)
		return nil
	})

	if !found {
		return nil, post.ErrNotFound
	}

	return &p, nil
}

func (r *PostRepository) List(_ context.Context, params post.ListParams) (*post.Page, error) {
	r.logger.Debug(
		"list posts",
		slog.Int("limit", params.Limit),
		slog.Bool("has_after", params.After != nil),
	)

	var posts []domain.Post
	_ = r.db.View(func(tx *inmem.Tx) error {
		ids := tx.PostIDs()
		start := 0
		if params.After != nil {
			start = sort.Search(len(ids), func(i int) bool {
				p, _ := tx.Post(ids[i])
				return positionBefore(
					p.CreatedAt,
					p.ID,
					params.After.CreatedAt,
					params.After.ID,
				)
			})
		}

		end := min(start+params.Limit+1, len(ids))
		posts = make([]domain.Post, 0, end-start)
		for _, id := range ids[start:end] {
			p, _ := tx.Post(id)
			posts = append(posts, p)
		}

		return nil
	})

	return makePostPage(posts, params.Limit), nil
}

func (r *PostRepository) SetCommentsEnabled(
	_ context.Context,
	postID uuid.UUID,
	authorID uuid.UUID,
	enabled bool,
) (*domain.Post, error) {
	r.logger.Debug(
		"set post comments enabled",
		slog.String("post_id", postID.String()),
		slog.String("author_id", authorID.String()),
		slog.Bool("enabled", enabled),
	)

	var p domain.Post
	err := r.db.Update(func(tx *inmem.Tx) error {
		found, ok := tx.Post(postID)
		if !ok {
			return post.ErrNotFound
		}
		if found.AuthorID != authorID {
			return post.ErrForbidden
		}

		found.CommentsEnabled = enabled
		found.UpdatedAt = time.Now().UTC()
		tx.PutPost(found)
		p = found
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func makePostPage(posts []domain.Post, limit int) *post.Page {
	hasNextPage := len(posts) > limit
	if hasNextPage {
		posts = posts[:limit]
	}

	page := &post.Page{
		Posts:       posts,
		HasNextPage: hasNextPage,
	}
	if len(posts) > 0 {
		last := posts[len(posts)-1]
		page.EndCursor = &post.Position{
			CreatedAt: last.CreatedAt,
			ID:        last.ID,
		}
	}

	return page
}

func positionBefore(createdAt time.Time, id uuid.UUID, cursorTime time.Time, cursorID uuid.UUID) bool {
	return createdAt.Before(cursorTime) ||
		(createdAt.Equal(cursorTime) && bytes.Compare(id[:], cursorID[:]) < 0)
}
