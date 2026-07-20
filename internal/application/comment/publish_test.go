package comment

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

	repo, notifier, svc := newTestService(t)
	postID, parentID, authorID := uuid.New(), uuid.New(), uuid.New()
	before := time.Now().UTC()
	repositoryResult := &domain.Comment{ID: uuid.New(), PostID: postID, AuthorID: authorID}

	var submitted *domain.Comment
	repo.EXPECT().Publish(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, comment *domain.Comment) (*domain.Comment, error) {
			submitted = comment
			return repositoryResult, nil
		},
	)
	notifier.EXPECT().NotifyCommentCreated(gomock.Any(), repositoryResult).DoAndReturn(
		func(_ context.Context, comment *domain.Comment) error {
			require.Same(t, repositoryResult, comment)
			return nil
		},
	)

	created, err := svc.Publish(t.Context(), PublishParams{
		PostID:   postID,
		ParentID: &parentID,
		AuthorID: authorID,
		Body:     "комментарий",
	})
	after := time.Now().UTC()

	require.NoError(t, err)
	require.Same(t, repositoryResult, created)
	require.NotEqual(t, uuid.Nil, submitted.ID)
	require.Equal(t, postID, submitted.PostID)
	require.NotNil(t, submitted.ParentID)
	require.Equal(t, parentID, *submitted.ParentID)
	require.Equal(t, authorID, submitted.AuthorID)
	require.Equal(t, "комментарий", submitted.Body)
	require.WithinRange(t, submitted.CreatedAt, before, after)
	require.Equal(t, time.UTC, submitted.CreatedAt.Location())
}

func TestService_Publish_RejectsInvalidComment(t *testing.T) {
	t.Parallel()

	_, _, svc := newTestService(t)
	created, err := svc.Publish(t.Context(), PublishParams{
		PostID:   uuid.New(),
		AuthorID: uuid.New(),
		Body:     "",
	})

	require.Nil(t, created)
	require.ErrorContains(t, err, "invalid comment")
	require.ErrorIs(t, err, domain.ErrEmptyComment)
}

func TestService_Publish_RepositoryFails(t *testing.T) {
	t.Parallel()

	repo, _, service := newTestService(t)
	repo.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil, errRepo)

	created, err := service.Publish(t.Context(), PublishParams{
		PostID:   uuid.New(),
		AuthorID: uuid.New(),
		Body:     "комментарий",
	})

	require.Nil(t, created)
	require.ErrorContains(t, err, "publish comment")
	require.ErrorIs(t, err, errRepo)
}

func TestService_Publish_IgnoresNotifierFail(t *testing.T) {
	t.Parallel()

	repo, notifier, svc := newTestService(t)
	var saved *domain.Comment
	repo.EXPECT().Publish(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, comment *domain.Comment) (*domain.Comment, error) {
			saved = comment
			return comment, nil
		},
	)
	notifier.EXPECT().NotifyCommentCreated(gomock.Any(), gomock.Any()).Return(errNotifier)

	created, err := svc.Publish(t.Context(), PublishParams{
		PostID:   uuid.New(),
		AuthorID: uuid.New(),
		Body:     "комментарий",
	})

	require.NoError(t, err)
	require.Same(t, saved, created)
}

func TestService_Publish_IsSavedBeforeNotifying(t *testing.T) {
	t.Parallel()

	repo, notifier, service := newTestService(t)
	gomock.InOrder(
		repo.EXPECT().Publish(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, comment *domain.Comment) (*domain.Comment, error) {
				return comment, nil
			},
		),
		notifier.EXPECT().NotifyCommentCreated(gomock.Any(), gomock.Any()).Return(nil),
	)

	_, err := service.Publish(t.Context(), PublishParams{
		PostID:   uuid.New(),
		AuthorID: uuid.New(),
		Body:     "комментарий",
	})

	require.NoError(t, err)
}
