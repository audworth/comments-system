package post

import (
	"testing"
	"time"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestService_PublishNewPost(t *testing.T) {
	t.Parallel()

	repo, svc := newTestService()
	authorID := uuid.New()
	before := time.Now().UTC()

	created, err := svc.PublishNewPost(t.Context(), NewPostParams{
		AuthorID:        authorID,
		Title:           "title",
		Body:            "body",
		CommentsEnabled: true,
	})
	after := time.Now().UTC()

	require.NoError(t, err)
	require.Same(t, repo.newPostInput, created)
	require.Equal(t, 1, repo.newPostCalls)
	require.NotEqual(t, uuid.Nil, repo.newPostInput.ID)
	require.Equal(t, domain.User{ID: authorID}, repo.newPostInput.Author)
	require.Equal(t, "title", repo.newPostInput.Title)
	require.Equal(t, "body", repo.newPostInput.Body)
	require.True(t, repo.newPostInput.CommentsEnabled)
	require.WithinRange(t, repo.newPostInput.CreatedAt, before, after)
	require.Equal(t, time.UTC, repo.newPostInput.CreatedAt.Location())
}

func TestService_PublishNewPost_RejectsInvalidPost(t *testing.T) {
	t.Parallel()

	repo, svc := newTestService()
	created, err := svc.PublishNewPost(t.Context(), NewPostParams{
		AuthorID: uuid.New(),
		Title:    "",
		Body:     "body",
	})

	require.Nil(t, created)
	require.ErrorContains(t, err, "invalid post")
	require.ErrorIs(t, err, domain.ErrEmptyPostTitle)
	require.Zero(t, repo.newPostCalls)
}

func TestService_PublishNewPost_ReturnsRepositoryResult(t *testing.T) {
	t.Parallel()

	repositoryResult := &domain.Post{
		ID:              uuid.New(),
		Author:          domain.User{ID: uuid.New()},
		Title:           "заголовок",
		Body:            "тело",
		CommentsEnabled: true,
		CreatedAt:       time.Now().UTC().Add(time.Second),
	}
	repo, svc := newTestService()
	repo.newPostResult = repositoryResult

	created, err := svc.PublishNewPost(t.Context(), NewPostParams{
		AuthorID: repositoryResult.Author.ID,
		Title:    "заголовок1",
		Body:     "тело1",
	})

	require.NoError(t, err)
	require.Same(t, repositoryResult, created)
	require.NotEqual(t, repo.newPostInput.ID, created.ID)
}

func TestService_PublishNewPost_RepositoryFails(t *testing.T) {
	t.Parallel()

	repo, svc := newTestService()
	repo.newPostErr = ErrNotFound

	created, err := svc.PublishNewPost(t.Context(), NewPostParams{
		AuthorID: uuid.New(),
		Title:    "заголовок",
		Body:     "тело",
	})

	require.Nil(t, created)
	require.ErrorContains(t, err, "publish post")
	require.ErrorIs(t, err, repo.newPostErr)
	require.Equal(t, 1, repo.newPostCalls)
}
