package post

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_GetByID_RepositoryFail(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	repo, svc := newTestService(t)
	repo.EXPECT().GetByID(gomock.Any(), id).Return(nil, ErrNotFound)

	got, err := svc.GetByID(t.Context(), id)

	require.Nil(t, got)
	require.ErrorContains(t, err, "get post "+id.String())
	require.ErrorIs(t, err, ErrNotFound)
}
