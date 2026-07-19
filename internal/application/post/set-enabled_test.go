package post

import (
	"testing"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_SetCommentsToEnabled(t *testing.T) {
	t.Parallel()

	postID, authorID := uuid.New(), uuid.New()
	want := &domain.Post{ID: postID, AuthorID: authorID, CommentsEnabled: true}
	repo, svc := newTestService(t)
	repo.EXPECT().SetCommentsEnabled(gomock.Any(), postID, authorID, true).Return(want, nil)

	got, err := svc.SetCommentsToEnabled(t.Context(), postID, authorID, true)

	require.NoError(t, err)
	require.Same(t, want, got)
}

func TestService_SetCommentsToEnabled_RepositoryFails(t *testing.T) {
	t.Parallel()

	postID, authorID := uuid.New(), uuid.New()
	repo, svc := newTestService(t)
	repo.EXPECT().SetCommentsEnabled(gomock.Any(), postID, authorID, true).Return(nil, ErrForbidden)

	got, err := svc.SetCommentsToEnabled(t.Context(), postID, authorID, true)

	require.Nil(t, got)
	require.ErrorContains(t, err, "set comments enabled for post "+postID.String())
	require.ErrorContains(t, err, "author "+authorID.String())
	require.ErrorIs(t, err, ErrForbidden)
}
