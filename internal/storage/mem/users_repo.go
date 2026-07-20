package mem

import (
	"context"
	"log/slog"

	"github.com/audworth/comments-system/internal/application/user"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/audworth/comments-system/internal/platform/db/inmem"
	"github.com/google/uuid"
)

var _ user.Repository = (*UserRepository)(nil)

type UserRepository struct {
	db     *inmem.InMem
	logger *slog.Logger
}

func NewUserRepository(db *inmem.InMem, logger *slog.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

func (r *UserRepository) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	r.logger.Debug("get user", slog.String("user_id", id.String()))

	var (
		u     domain.User
		found bool
	)
	_ = r.db.View(func(tx *inmem.Tx) error {
		u, found = tx.User(id)
		return nil
	})

	if !found {
		return nil, user.ErrNotFound
	}

	return &u, nil
}

func (r *UserRepository) GetByIDs(_ context.Context, ids []uuid.UUID) (map[uuid.UUID]*domain.User, error) {
	r.logger.Debug("get users", slog.Int("user_count", len(ids)))

	users := make(map[uuid.UUID]*domain.User, len(ids))
	_ = r.db.View(func(tx *inmem.Tx) error {
		for _, id := range ids {
			u, ok := tx.User(id)
			if !ok {
				continue
			}

			value := u
			users[id] = &value
		}
		return nil
	})

	return users, nil
}
