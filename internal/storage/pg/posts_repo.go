package pg

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/audworth/comments-system/internal/application/post"
	"github.com/audworth/comments-system/internal/application/user"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ post.Repository = (*PostRepository)(nil)

type PostRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewPostRepository(db *pgxpool.Pool, logger *slog.Logger) *PostRepository {
	return &PostRepository{db: db, logger: logger}
}

func (r *PostRepository) Publish(ctx context.Context, post *domain.Post) (*domain.Post, error) {
	row := r.db.QueryRow(ctx, `
		insert into posts (
			id,
			author_id,
			title,
			body,
			comments_enabled,
			created_at,
			updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7)
		returning
			id,
			author_id,
			title,
			body,
			comments_enabled,
			created_at,
			updated_at
	`,
		post.ID,
		post.AuthorID,
		post.Title,
		post.Body,
		post.CommentsEnabled,
		post.CreatedAt,
		post.UpdatedAt,
	)

	p := &domain.Post{}
	err := row.Scan(
		&p.ID,
		&p.AuthorID,
		&p.Title,
		&p.Body,
		&p.CommentsEnabled,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if isForeignKeyViolation(err, postsAuthorConstraint) {
			return nil, user.ErrNotFound
		}
		r.logger.ErrorContext(
			ctx,
			"failed to create post",
			slog.String("post_id", post.ID.String()),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("publish post: %w", err)
	}

	return p, nil
}

func (r *PostRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
	row := r.db.QueryRow(ctx, `
		select
			id,
			author_id,
			title,
			body,
			comments_enabled,
			created_at,
			updated_at
		from posts
		where id = $1
	`, id)

	p := &domain.Post{}
	err := row.Scan(
		&p.ID,
		&p.AuthorID,
		&p.Title,
		&p.Body,
		&p.CommentsEnabled,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, post.ErrNotFound
		}
		r.logger.ErrorContext(
			ctx,
			"failed to get post",
			slog.String("post_id", id.String()),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("get post %s: %w", id, err)
	}

	return p, nil
}

func (r *PostRepository) List(ctx context.Context, params post.ListParams) (*post.Page, error) {
	query := `
		select
			id,
			author_id,
			title,
			body,
			comments_enabled,
			created_at,
			updated_at
		from posts
	`

	var args []any
	if params.After == nil {
		query += `order by created_at desc, id desc limit $1`
		args = []any{params.Limit + 1}
	} else {
		query += `
			where (created_at, id) < ($1, $2)
			order by created_at desc, id desc
			limit $3
		`
		args = []any{
			params.After.CreatedAt,
			params.After.ID,
			params.Limit + 1,
		}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to list posts", slog.Any("error", err))
		return nil, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()

	posts := make([]domain.Post, 0, params.Limit+1)
	for rows.Next() {
		p := &domain.Post{}
		err := rows.Scan(
			&p.ID,
			&p.AuthorID,
			&p.Title,
			&p.Body,
			&p.CommentsEnabled,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}

		posts = append(posts, *p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate posts: %w", err)
	}

	hasNextPage := len(posts) > params.Limit
	if hasNextPage {
		posts = posts[:params.Limit]
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

	return page, nil
}

func (r *PostRepository) SetCommentsEnabled(
	ctx context.Context,
	postID uuid.UUID,
	authorID uuid.UUID,
	enabled bool,
) (*domain.Post, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.logger.ErrorContext(
			ctx,
			"failed to begin comments setting transaction",
			slog.String("post_id", postID.String()),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("set comments enabled transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var foundAuthorID uuid.UUID
	err = tx.QueryRow(ctx, `
		select author_id
		from posts
		where id = $1
		for update
	`, postID).Scan(&foundAuthorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, post.ErrNotFound
		}
		r.logger.ErrorContext(
			ctx,
			"failed to lock post",
			slog.String("post_id", postID.String()),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("lock post %s: %w", postID, err)
	}

	if foundAuthorID != authorID {
		return nil, post.ErrForbidden
	}

	row := tx.QueryRow(ctx, `
		update posts
		set
			comments_enabled = $2,
			updated_at = now()
		where id = $1
		returning
			id,
			author_id,
			title,
			body,
			comments_enabled,
			created_at,
			updated_at
	`, postID, enabled)

	result := &domain.Post{}
	if err := row.Scan(
		&result.ID,
		&result.AuthorID,
		&result.Title,
		&result.Body,
		&result.CommentsEnabled,
		&result.CreatedAt,
		&result.UpdatedAt,
	); err != nil {
		r.logger.ErrorContext(
			ctx,
			"failed to update post comments setting",
			slog.String("post_id", postID.String()),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("update post %s: %w", postID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		r.logger.ErrorContext(
			ctx,
			"failed to commit comments setting transaction",
			slog.String("post_id", postID.String()),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("commit set comments for post %s: %w", postID, err)
	}

	return result, nil
}
