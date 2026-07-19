package post

import (
	"testing"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_GetByID(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	want := &domain.Post{ID: id}
	repo, svc := newTestService(t)
	repo.EXPECT().PostByID(gomock.Any(), id).Return(want, nil)

	got, err := svc.GetByID(t.Context(), id)

	require.NoError(t, err)
	require.Same(t, want, got)
}

func TestService_GetByID_RepositoryFail(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	repo, svc := newTestService(t)
	repo.EXPECT().PostByID(gomock.Any(), id).Return(nil, ErrNotFound)

	got, err := svc.GetByID(t.Context(), id)

	require.Nil(t, got)
	require.ErrorContains(t, err, "get post "+id.String())
	require.ErrorIs(t, err, ErrNotFound)
}
