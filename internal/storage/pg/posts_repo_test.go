package pg

import (
	"testing"
	"time"

	"github.com/audworth/comments-system/internal/application/post"
	"github.com/audworth/comments-system/internal/application/user"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPostRepositoryIntegration_LifecycleAndPagination(t *testing.T) {
	t.Parallel()

	fixture := newPGFixture(t)
	oldest := publishTestPost(t, fixture, testID(10), true, testTime.Add(-time.Second))
	middle := publishTestPost(t, fixture, testID(11), true, testTime)
	newest := publishTestPost(t, fixture, testID(12), true, testTime)

	stored, err := fixture.posts.GetByID(t.Context(), middle.ID)
	require.NoError(t, err)
	require.Equal(t, middle, *stored)

	first, err := fixture.posts.List(t.Context(), post.ListParams{Limit: 2})
	require.NoError(t, err)
	require.Equal(t, []domain.Post{newest, middle}, first.Posts)
	require.True(t, first.HasNextPage)
	require.Equal(t, &post.Position{CreatedAt: middle.CreatedAt, ID: middle.ID}, first.EndCursor)

	second, err := fixture.posts.List(t.Context(), post.ListParams{Limit: 2, After: first.EndCursor})
	require.NoError(t, err)
	require.Equal(t, []domain.Post{oldest}, second.Posts)
	require.False(t, second.HasNextPage)

	updated, err := fixture.posts.SetCommentsEnabled(t.Context(), middle.ID, uuid.New(), false)
	require.Nil(t, updated)
	require.ErrorIs(t, err, post.ErrForbidden)

	updated, err = fixture.posts.SetCommentsEnabled(t.Context(), middle.ID, fixture.user.ID, false)
	require.NoError(t, err)
	require.False(t, updated.CommentsEnabled)
	require.NotEqual(t, middle.UpdatedAt, updated.UpdatedAt)

	missing, err := fixture.posts.GetByID(t.Context(), uuid.New())
	require.Nil(t, missing)
	require.ErrorIs(t, err, post.ErrNotFound)
}

func TestPostRepositoryIntegration_PublishRejectsUnknownAuthor(t *testing.T) {
	t.Parallel()

	fixture := newPGFixture(t)
	p := domain.Post{
		ID:              testID(10),
		AuthorID:        uuid.New(),
		Title:           "post",
		Body:            "body",
		CommentsEnabled: true,
		CreatedAt:       testTime,
		UpdatedAt:       testTime,
	}

	created, err := fixture.posts.Publish(t.Context(), &p)
	require.Nil(t, created)
	require.ErrorIs(t, err, user.ErrNotFound)
}
