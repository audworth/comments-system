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

func TestService_PublishNewComment(t *testing.T) {
	t.Parallel()

	repo, notifier, svc := newTestService(t)
	postID, parentID, authorID := uuid.New(), uuid.New(), uuid.New()
	before := time.Now().UTC()

	var saved *domain.Comment
	repo.EXPECT().NewComment(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, comment *domain.Comment) (*domain.Comment, error) {
			saved = comment
			return comment, nil
		},
	)
	notifier.EXPECT().NotifyCreated(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, comment *domain.Comment) error {
			require.Same(t, saved, comment)
			return nil
		},
	)

	created, err := svc.PublishNewComment(t.Context(), &NewCommentParams{
		PostID:   postID,
		ParentID: &parentID,
		AuthorID: authorID,
		Body:     "комментарий",
	})
	after := time.Now().UTC()

	require.NoError(t, err)
	require.Same(t, saved, created)
	require.NotEqual(t, uuid.Nil, saved.ID)
	require.Equal(t, postID, saved.PostID)
	require.NotNil(t, saved.ParentID)
	require.Equal(t, parentID, *saved.ParentID)
	require.Equal(t, authorID, saved.AuthorID)
	require.Equal(t, "комментарий", saved.Body)
	require.WithinRange(t, saved.CreatedAt, before, after)
	require.Equal(t, time.UTC, saved.CreatedAt.Location())
}

func TestService_PublishNewComment_RejectsInvalidComment(t *testing.T) {
	t.Parallel()

	_, _, svc := newTestService(t)
	created, err := svc.PublishNewComment(t.Context(), &NewCommentParams{
		PostID:   uuid.New(),
		AuthorID: uuid.New(),
		Body:     "",
	})

	require.Nil(t, created)
	require.ErrorContains(t, err, "invalid comment")
	require.ErrorIs(t, err, domain.ErrEmptyComment)
}

func TestService_PublishNewComment_ReturnsAndNotifiesRepositoryResult(t *testing.T) {
	t.Parallel()

	repoResult := &domain.Comment{
		ID:        uuid.New(),
		PostID:    uuid.New(),
		AuthorID:  uuid.New(),
		Body:      "сохраненный",
		CreatedAt: time.Now().UTC().Add(time.Second),
	}
	repo, notifier, svc := newTestService(t)
	var input *domain.Comment
	repo.EXPECT().NewComment(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, comment *domain.Comment) (*domain.Comment, error) {
			input = comment
			return repoResult, nil
		},
	)
	notifier.EXPECT().NotifyCreated(gomock.Any(), repoResult).Return(nil)

	created, err := svc.PublishNewComment(t.Context(), &NewCommentParams{
		PostID:   repoResult.PostID,
		AuthorID: repoResult.AuthorID,
		Body:     "исходный",
	})

	require.NoError(t, err)
	require.Same(t, repoResult, created)
	require.NotEqual(t, input.ID, created.ID)
}

func TestService_PublishNewComment_RepositoryFails(t *testing.T) {
	t.Parallel()

	repo, _, service := newTestService(t)
	repo.EXPECT().NewComment(gomock.Any(), gomock.Any()).Return(nil, errRepo)

	created, err := service.PublishNewComment(t.Context(), &NewCommentParams{
		PostID:   uuid.New(),
		AuthorID: uuid.New(),
		Body:     "комментарий",
	})

	require.Nil(t, created)
	require.ErrorContains(t, err, "publish comment")
	require.ErrorIs(t, err, errRepo)
}

func TestService_PublishNewComment_IgnoresNotifierFail(t *testing.T) {
	t.Parallel()

	repo, notifier, svc := newTestService(t)
	var saved *domain.Comment
	repo.EXPECT().NewComment(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, comment *domain.Comment) (*domain.Comment, error) {
			saved = comment
			return comment, nil
		},
	)
	notifier.EXPECT().NotifyCreated(gomock.Any(), gomock.Any()).Return(errNotifier)

	created, err := svc.PublishNewComment(t.Context(), &NewCommentParams{
		PostID:   uuid.New(),
		AuthorID: uuid.New(),
		Body:     "комментарий",
	})

	require.NoError(t, err)
	require.Same(t, saved, created)
}

func TestService_PublishNewComment_IsSavedBeforeNotifying(t *testing.T) {
	t.Parallel()

	repo, notifier, service := newTestService(t)
	gomock.InOrder(
		repo.EXPECT().NewComment(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, comment *domain.Comment) (*domain.Comment, error) {
				return comment, nil
			},
		),
		notifier.EXPECT().NotifyCreated(gomock.Any(), gomock.Any()).Return(nil),
	)

	_, err := service.PublishNewComment(t.Context(), &NewCommentParams{
		PostID:   uuid.New(),
		AuthorID: uuid.New(),
		Body:     "комментарий",
	})

	require.NoError(t, err)
}
