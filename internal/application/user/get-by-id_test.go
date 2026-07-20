package user

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var errRepo = errors.New("db unavailable")

func newTestService(t *testing.T) (*MockRepository, *Service) {
	t.Helper()

	repo := NewMockRepository(gomock.NewController(t))
	return repo, NewService(repo, slog.New(slog.DiscardHandler))
}

func TestService_GetByID_RepositoryFail(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	repo, svc := newTestService(t)
	repo.EXPECT().GetByID(gomock.Any(), id).Return(nil, errRepo)

	got, err := svc.GetByID(t.Context(), id)

	require.Nil(t, got)
	require.ErrorContains(t, err, "get user "+id.String())
	require.ErrorIs(t, err, errRepo)
}

func TestService_GetByIDs_RepositoryFails(t *testing.T) {
	t.Parallel()

	ids := []uuid.UUID{uuid.New(), uuid.New()}
	repo, svc := newTestService(t)
	repo.EXPECT().GetByIDs(gomock.Any(), ids).Return(nil, errRepo)

	users, err := svc.GetByIDs(t.Context(), ids)

	require.Nil(t, users)
	require.ErrorContains(t, err, "get users")
	require.ErrorIs(t, err, errRepo)
}
