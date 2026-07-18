package post

import (
	"testing"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestService_SetCommentsToEnabled(t *testing.T) {
	t.Parallel()

	names := map[bool]string{false: "disabled", true: "enabled"}
	for _, enabled := range []bool{false, true} {
		t.Run(names[enabled], func(t *testing.T) {
			t.Parallel()

			postID, authorID := uuid.New(), uuid.New()
			want := &domain.Post{ID: postID, Author: domain.User{ID: authorID}, CommentsEnabled: enabled}
			repo, svc := newTestService()
			repo.setCommentsEnabledResult = want

			got, err := svc.SetCommentsToEnabled(t.Context(), postID, authorID, enabled)

			require.NoError(t, err)
			require.Same(t, want, got)
			require.Equal(t, 1, repo.setCommentsEnabledCalls)
			require.Equal(t, postID, repo.setCommentsEnabledPostID)
			require.Equal(t, authorID, repo.setCommentsEnabledAuthor)
			require.Equal(t, enabled, repo.setCommentsEnabledEnabled)
		})
	}
}

func TestService_SetCommentsToEnabled_RepositoryFails(t *testing.T) {
	t.Parallel()

	postID, authorID := uuid.New(), uuid.New()
	repo, svc := newTestService()
	repo.setCommentsEnabledErr = ErrForbidden

	got, err := svc.SetCommentsToEnabled(t.Context(), postID, authorID, true)

	require.Nil(t, got)
	require.ErrorContains(t, err, "set comments enabled for post "+postID.String())
	require.ErrorContains(t, err, "author "+authorID.String())
	require.ErrorIs(t, err, repo.setCommentsEnabledErr)
	require.Equal(t, 1, repo.setCommentsEnabledCalls)
	require.Equal(t, postID, repo.setCommentsEnabledPostID)
	require.Equal(t, authorID, repo.setCommentsEnabledAuthor)
	require.True(t, repo.setCommentsEnabledEnabled)
}
