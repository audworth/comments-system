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

func TestService_Publish(t *testing.T) {
	t.Parallel()

	repo, svc := newTestService(t)
	authorID := uuid.New()
	before := time.Now().UTC()
	repositoryResult := &domain.Post{ID: uuid.New(), AuthorID: authorID}
	var submitted *domain.Post
	repo.EXPECT().Publish(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, post *domain.Post) (*domain.Post, error) {
			submitted = post
			return repositoryResult, nil
		},
	)

	created, err := svc.Publish(t.Context(), PublishParams{
		AuthorID:        authorID,
		Title:           "title",
		Body:            "body",
		CommentsEnabled: true,
	})
	after := time.Now().UTC()

	require.NoError(t, err)
	require.Same(t, repositoryResult, created)
	require.NotEqual(t, uuid.Nil, submitted.ID)
	require.Equal(t, authorID, submitted.AuthorID)
	require.Equal(t, "title", submitted.Title)
	require.Equal(t, "body", submitted.Body)
	require.True(t, submitted.CommentsEnabled)
	require.WithinRange(t, submitted.CreatedAt, before, after)
	require.Equal(t, time.UTC, submitted.CreatedAt.Location())
}

func TestService_Publish_RejectsInvalidPost(t *testing.T) {
	t.Parallel()

	_, svc := newTestService(t)
	created, err := svc.Publish(t.Context(), PublishParams{
		AuthorID: uuid.New(),
		Title:    "",
		Body:     "body",
	})

	require.Nil(t, created)
	require.ErrorContains(t, err, "invalid post")
	require.ErrorIs(t, err, domain.ErrEmptyPostTitle)
}

func TestService_Publish_RepositoryFails(t *testing.T) {
	t.Parallel()

	repo, svc := newTestService(t)
	repo.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil, ErrNotFound)

	created, err := svc.Publish(t.Context(), PublishParams{
		AuthorID: uuid.New(),
		Title:    "заголовок",
		Body:     "тело",
	})

	require.Nil(t, created)
	require.ErrorContains(t, err, "publish post")
	require.ErrorIs(t, err, ErrNotFound)
}
