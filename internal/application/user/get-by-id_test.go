package user

import (
	"errors"
	"testing"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var errRepo = errors.New("db unavailable")

func newTestService(t *testing.T) (*MockRepository, *Service) {
	t.Helper()

	repo := NewMockRepository(gomock.NewController(t))
	return repo, NewService(repo)
}

func TestService_GetByID(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	want := &domain.User{ID: id, Name: "Дмитрий"}
	repo, svc := newTestService(t)
	repo.EXPECT().UserByID(gomock.Any(), id).Return(want, nil)

	got, err := svc.GetByID(t.Context(), id)

	require.NoError(t, err)
	require.Same(t, want, got)
}

func TestService_GetByID_RepositoryFails(t *testing.T) {
	t.Parallel()

	repoErr := errRepo
	id := uuid.New()
	repo, svc := newTestService(t)
	repo.EXPECT().UserByID(gomock.Any(), id).Return(nil, repoErr)

	got, err := svc.GetByID(t.Context(), id)

	require.Nil(t, got)
	require.ErrorContains(t, err, "get user "+id.String())
	require.ErrorIs(t, err, repoErr)
}
