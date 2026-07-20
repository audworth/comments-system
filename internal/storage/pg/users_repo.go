package pg

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/audworth/comments-system/internal/application/user"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ user.Repository = (*UserRepository)(nil)

type UserRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewUserRepository(db *pgxpool.Pool, logger *slog.Logger) *UserRepository {
	return &UserRepository{db: db, logger: logger}
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row := r.db.QueryRow(ctx, `
		select id, name
		from users
		where id = $1
	`, id)

	u := &domain.User{}
	err := row.Scan(
		&u.ID,
		&u.Name,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		r.logger.ErrorContext(
			ctx,
			"failed to get user",
			slog.String("user_id", id.String()),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("get user %s: %w", id, err)
	}

	return u, nil
}

func (r *UserRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domain.User, error) {
	users := make(map[uuid.UUID]*domain.User, len(ids))
	if len(ids) == 0 {
		return users, nil
	}

	rows, err := r.db.Query(ctx, `
		select id, name
		from users
		where id = any($1)
	`, ids)
	if err != nil {
		r.logger.ErrorContext(
			ctx,
			"failed to get users",
			slog.Int("user_count", len(ids)),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("get users by ids: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		u := &domain.User{}
		err := rows.Scan(
			&u.ID,
			&u.Name,
		)

		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}

		users[u.ID] = u
	}
	if err := rows.Err(); err != nil {
		r.logger.ErrorContext(ctx, "failed while iterating user rows", slog.Any("error", err))
		return nil, fmt.Errorf("get users by ids: %w", err)
	}

	return users, nil
}
