package mem

import (
	"log/slog"
	"testing"

	"github.com/audworth/comments-system/internal/application/user"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_GetByID(t *testing.T) {
	t.Parallel()

	want := domain.User{ID: testID(1), Name: "user_1"}
	repo := NewUserRepository(
		newTestDB(t, []domain.User{want}, nil, nil),
		slog.New(slog.DiscardHandler),
	)

	got, err := repo.GetByID(t.Context(), want.ID)

	require.NoError(t, err)
	require.Equal(t, &want, got)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	repo := NewUserRepository(newTestDB(t, nil, nil, nil), slog.New(slog.DiscardHandler))

	got, err := repo.GetByID(t.Context(), testID(1))

	require.Nil(t, got)
	require.ErrorIs(t, err, user.ErrNotFound)
}

func TestUserRepository_GetByIDs(t *testing.T) {
	t.Parallel()

	u1 := domain.User{ID: testID(1), Name: "user_1"}
	u2 := domain.User{ID: testID(2), Name: "user_2"}
	missingID := testID(3)
	repo := NewUserRepository(
		newTestDB(t, []domain.User{u1, u2}, nil, nil),
		slog.New(slog.DiscardHandler),
	)

	got, err := repo.GetByIDs(t.Context(), []uuid.UUID{u2.ID, missingID, u1.ID, u1.ID})

	require.NoError(t, err)
	require.Equal(t, map[uuid.UUID]*domain.User{
		u1.ID: &u1,
		u2.ID: &u2,
	}, got)
}
