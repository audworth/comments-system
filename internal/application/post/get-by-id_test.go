package post

import (
	"testing"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestService_PostByID(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	want := &domain.Post{ID: id}
	repo, svc := newTestService()
	repo.postByIDResult = want

	got, err := svc.PostByID(t.Context(), id)

	require.NoError(t, err)
	require.Same(t, want, got)
	require.Equal(t, 1, repo.postByIDCalls)
	require.Equal(t, id, repo.postByIDInput)
}

func TestService_PostByID_RepositoryFail(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	repo, svc := newTestService()
	repo.postByIDErr = ErrNotFound

	got, err := svc.PostByID(t.Context(), id)

	require.Nil(t, got)
	require.ErrorContains(t, err, "get post "+id.String())
	require.ErrorIs(t, err, repo.postByIDErr)
	require.Equal(t, 1, repo.postByIDCalls)
	require.Equal(t, id, repo.postByIDInput)
}
