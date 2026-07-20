package pg

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/audworth/comments-system/internal/application/comment"
	"github.com/audworth/comments-system/internal/application/user"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ comment.Repository = (*CommentsRepository)(nil)

type CommentsRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewCommentsRepository(db *pgxpool.Pool, logger *slog.Logger) *CommentsRepository {
	return &CommentsRepository{db: db, logger: logger}
}

func (r *CommentsRepository) Publish(ctx context.Context, newComm *domain.Comment) (*domain.Comment, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.logger.ErrorContext(
			ctx,
			"failed to begin transaction for new comment",
			slog.String("post_id", newComm.PostID.String()),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("publish comment transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var commentsEnabled bool
	err = tx.QueryRow(ctx, `
		select comments_enabled
		from posts
		where id = $1
		for share
	`, newComm.PostID).Scan(&commentsEnabled)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, comment.ErrPostNotFound
		}
		r.logger.ErrorContext(
			ctx,
			"failed to lock post for comment",
			slog.String("post_id", newComm.PostID.String()),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("lock post %s: %w", newComm.PostID, err)
	}
	if !commentsEnabled {
		return nil, comment.ErrCommentsDisabled
	}

	if newComm.ParentID != nil {
		var parentID uuid.UUID
		err = tx.QueryRow(ctx, `
			select id
			from comments
			where post_id = $1 and id = $2
		`, newComm.PostID, *newComm.ParentID).Scan(&parentID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, comment.ErrParentNotFound
			}
			r.logger.ErrorContext(
				ctx,
				"failed to check parent comment",
				slog.String("post_id", newComm.PostID.String()),
				slog.String("parent_id", newComm.ParentID.String()),
				slog.Any("error", err),
			)
			return nil, fmt.Errorf("check parent comment %s: %w", *newComm.ParentID, err)
		}
	}

	row := tx.QueryRow(ctx, `
		insert into comments (
			id,
			post_id,
			parent_id,
			author_id,
			body,
			created_at
		)
		values ($1, $2, $3, $4, $5, $6)
		returning
			id,
			post_id,
			parent_id,
			author_id,
			body,
			created_at
	`,
		newComm.ID,
		newComm.PostID,
		newComm.ParentID,
		newComm.AuthorID,
		newComm.Body,
		newComm.CreatedAt,
	)

	comm := &domain.Comment{}
	err = row.Scan(
		&comm.ID,
		&comm.PostID,
		&comm.ParentID,
		&comm.AuthorID,
		&comm.Body,
		&comm.CreatedAt,
	)
	if err != nil {
		if isForeignKeyViolation(err, commentsAuthorConstraint) {
			return nil, user.ErrNotFound
		}
		r.logger.ErrorContext(
			ctx,
			"failed to create comment",
			slog.String("comment_id", newComm.ID.String()),
			slog.String("post_id", newComm.PostID.String()),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("publish comment: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		r.logger.ErrorContext(
			ctx,
			"failed to commit transaction",
			slog.String("comment_id", comm.ID.String()),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("commit comment %s: %w", comm.ID, err)
	}

	return comm, nil
}

func (r *CommentsRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error) {
	row := r.db.QueryRow(ctx, `
		select
			id,
			post_id,
			parent_id,
			author_id,
			body,
			created_at
		from comments
		where id = $1
	`, id)

	comm := &domain.Comment{}
	err := row.Scan(
		&comm.ID,
		&comm.PostID,
		&comm.ParentID,
		&comm.AuthorID,
		&comm.Body,
		&comm.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, comment.ErrNotFound
		}
		r.logger.ErrorContext(
			ctx,
			"failed to get commen",
			slog.String("comment_id", id.String()),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("get comment %s: %w", id, err)
	}

	return comm, nil
}

func (r *CommentsRepository) List(ctx context.Context, params comment.ListParams) (*comment.Page, error) {
	query, args := makeListQuery(&params)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.logger.ErrorContext(
			ctx,
			"failed to list comments",
			slog.String("post_id", params.PostID.String()),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("list children for post %s: %w", params.PostID, err)
	}

	page, err := pageFromRows(rows, params.Limit)
	if err != nil {
		r.logger.ErrorContext(
			ctx,
			"failed to read comment page",
			slog.String("post_id", params.PostID.String()),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("list children for post %s: %w", params.PostID, err)
	}

	return page, nil
}

func (r *CommentsRepository) ListBatch(ctx context.Context, params []comment.ListParams) ([]*comment.Page, error) {
	pages := make([]*comment.Page, len(params))
	if len(params) == 0 {
		return pages, nil
	}

	batch := &pgx.Batch{}
	for i := range params {
		query, args := makeListQuery(&params[i])
		batch.Queue(query, args...)
	}

	results := r.db.SendBatch(ctx, batch)
	for i := range params {
		rows, err := results.Query()
		if err != nil {
			_ = results.Close()
			r.logger.ErrorContext(
				ctx,
				"failed to read comment batch",
				slog.Int("batch_index", i),
				slog.Any("error", err),
			)
			return nil, fmt.Errorf("list children batch item %d: %w", i, err)
		}

		page, err := pageFromRows(rows, params[i].Limit)
		if err != nil {
			_ = results.Close()
			r.logger.ErrorContext(
				ctx,
				"failed to read comment batch page",
				slog.Int("batch_index", i),
				slog.Any("error", err),
			)
			return nil, fmt.Errorf("list children batch item %d: %w", i, err)
		}
		pages[i] = page
	}

	if err := results.Close(); err != nil {
		r.logger.ErrorContext(ctx, "failed to close comment batch", slog.Any("error", err))
		return nil, fmt.Errorf("close children batch: %w", err)
	}

	return pages, nil
}

func makeListQuery(params *comment.ListParams) (string, []any) {
	query := `
		select
			id,
			post_id,
			parent_id,
			author_id,
			body,
			created_at
		from comments
		where post_id = $1
	`
	args := []any{params.PostID}
	if params.ParentID == nil {
		query += `and parent_id is null`
	} else {
		query += `and parent_id = $2`
		args = append(args, *params.ParentID)
	}

	if params.After != nil {
		afterCreatedAt := len(args) + 1
		afterID := afterCreatedAt + 1
		limit := afterID + 1
		query += `
			and (created_at, id) < ($%d, $%d)
			order by created_at desc, id desc
			limit $%d
		`
		query = fmt.Sprintf(query, afterCreatedAt, afterID, limit)
		args = append(args, params.After.CreatedAt, params.After.ID, params.Limit+1)
	} else {
		limitArg := len(args) + 1
		query += `
			order by created_at desc, id desc
			limit $%d
		`
		query = fmt.Sprintf(query, limitArg)
		args = append(args, params.Limit+1)
	}

	return query, args
}

func pageFromRows(rows pgx.Rows, limit int) (*comment.Page, error) {
	defer rows.Close()

	comments := make([]domain.Comment, 0, limit+1)
	for rows.Next() {
		comm := domain.Comment{}
		err := rows.Scan(
			&comm.ID,
			&comm.PostID,
			&comm.ParentID,
			&comm.AuthorID,
			&comm.Body,
			&comm.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}
		comments = append(comments, comm)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	hasNextPage := len(comments) > limit
	if hasNextPage {
		comments = comments[:limit]
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

	return page, nil
}
