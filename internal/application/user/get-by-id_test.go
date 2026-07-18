package user

import (
	"testing"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestService_UserByID(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	want := &domain.User{ID: id, Name: "Дмитрий"}
	repo, svc := newTestService()
	repo.userByIDResult = want

	got, err := svc.UserByID(t.Context(), id)

	require.NoError(t, err)
	require.Same(t, want, got)
	require.Equal(t, 1, repo.userByIDCalls)
	require.Equal(t, id, repo.userByIDInput)
}

func TestService_UserByID_RepositoryFails(t *testing.T) {
	t.Parallel()

	repoErr := errRepo
	id := uuid.New()
	repo, svc := newTestService()
	repo.userByIDErr = repoErr

	got, err := svc.UserByID(t.Context(), id)

	require.Nil(t, got)
	require.ErrorContains(t, err, "get user "+id.String())
	require.ErrorIs(t, err, repoErr)
	require.Equal(t, 1, repo.userByIDCalls)
	require.Equal(t, id, repo.userByIDInput)
}
