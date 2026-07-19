package post

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_SetCommentsEnabled_RepositoryFails(t *testing.T) {
	t.Parallel()

	postID, authorID := uuid.New(), uuid.New()
	repo, svc := newTestService(t)
	repo.EXPECT().SetCommentsEnabled(gomock.Any(), postID, authorID, true).Return(nil, ErrForbidden)

	got, err := svc.SetCommentsEnabled(t.Context(), postID, authorID, true)

	require.Nil(t, got)
	require.ErrorContains(t, err, "set comments enabled for post "+postID.String())
	require.ErrorContains(t, err, "author "+authorID.String())
	require.ErrorIs(t, err, ErrForbidden)
}
