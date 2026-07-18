package post

import (
	"context"
	"testing"
	"time"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_PublishNewPost(t *testing.T) {
	t.Parallel()

	repo, svc := newTestService(t)
	authorID := uuid.New()
	before := time.Now().UTC()
	var saved *domain.Post
	repo.EXPECT().NewPost(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, post *domain.Post) (*domain.Post, error) {
			saved = post
			return post, nil
		},
	)

	created, err := svc.PublishNewPost(t.Context(), NewPostParams{
		AuthorID:        authorID,
		Title:           "title",
		Body:            "body",
		CommentsEnabled: true,
	})
	after := time.Now().UTC()

	require.NoError(t, err)
	require.Same(t, saved, created)
	require.NotEqual(t, uuid.Nil, saved.ID)
	require.Equal(t, domain.User{ID: authorID}, saved.Author)
	require.Equal(t, "title", saved.Title)
	require.Equal(t, "body", saved.Body)
	require.True(t, saved.CommentsEnabled)
	require.WithinRange(t, saved.CreatedAt, before, after)
	require.Equal(t, time.UTC, saved.CreatedAt.Location())
}

func TestService_PublishNewPost_RejectsInvalidPost(t *testing.T) {
	t.Parallel()

	_, svc := newTestService(t)
	created, err := svc.PublishNewPost(t.Context(), NewPostParams{
		AuthorID: uuid.New(),
		Title:    "",
		Body:     "body",
	})

	require.Nil(t, created)
	require.ErrorContains(t, err, "invalid post")
	require.ErrorIs(t, err, domain.ErrEmptyPostTitle)
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
	repo, svc := newTestService(t)
	var saved *domain.Post
	repo.EXPECT().NewPost(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, post *domain.Post) (*domain.Post, error) {
			saved = post
			return repositoryResult, nil
		},
	)

	created, err := svc.PublishNewPost(t.Context(), NewPostParams{
		AuthorID: repositoryResult.Author.ID,
		Title:    "заголовок1",
		Body:     "тело1",
	})

	require.NoError(t, err)
	require.Same(t, repositoryResult, created)
	require.NotEqual(t, saved.ID, created.ID)
}

func TestService_PublishNewPost_RepositoryFails(t *testing.T) {
	t.Parallel()

	repo, svc := newTestService(t)
	repo.EXPECT().NewPost(gomock.Any(), gomock.Any()).Return(nil, ErrNotFound)

	created, err := svc.PublishNewPost(t.Context(), NewPostParams{
		AuthorID: uuid.New(),
		Title:    "заголовок",
		Body:     "тело",
	})

	require.Nil(t, created)
	require.ErrorContains(t, err, "publish post")
	require.ErrorIs(t, err, ErrNotFound)
}
