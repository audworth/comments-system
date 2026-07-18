package comment

import (
	"testing"
	"time"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestService_PublishNewComment(t *testing.T) {
	t.Parallel()

	repo, notifier, svc := newPublishTestService()
	postID, parentID, authorID := uuid.New(), uuid.New(), uuid.New()
	before := time.Now().UTC()

	created, err := svc.PublishNewComment(t.Context(), &NewCommentParams{
		PostID:   postID,
		ParentID: &parentID,
		AuthorID: authorID,
		Body:     "комментарий",
	})
	after := time.Now().UTC()

	require.NoError(t, err)
	require.Same(t, repo.newCommentInput, created)
	require.Equal(t, 1, repo.newCommentCalls)
	require.NotEqual(t, uuid.Nil, repo.newCommentInput.ID)
	require.Equal(t, postID, repo.newCommentInput.PostID)
	require.NotNil(t, repo.newCommentInput.ParentID)
	require.Equal(t, parentID, *repo.newCommentInput.ParentID)
	require.Equal(t, authorID, repo.newCommentInput.AuthorID)
	require.Equal(t, "комментарий", repo.newCommentInput.Body)
	require.WithinRange(t, repo.newCommentInput.CreatedAt, before, after)
	require.Equal(t, time.UTC, repo.newCommentInput.CreatedAt.Location())
	require.Equal(t, 1, notifier.notifyCreatedCalls)
	require.Same(t, created, notifier.notifyCreatedInput)
}

func TestService_PublishNewComment_RejectsInvalidComment(t *testing.T) {
	t.Parallel()

	repo, notifier, svc := newPublishTestService()
	created, err := svc.PublishNewComment(t.Context(), &NewCommentParams{
		PostID:   uuid.New(),
		AuthorID: uuid.New(),
		Body:     "",
	})

	require.Nil(t, created)
	require.ErrorContains(t, err, "invalid comment")
	require.ErrorIs(t, err, domain.ErrEmptyComment)
	require.Zero(t, repo.newCommentCalls)
	require.Zero(t, notifier.notifyCreatedCalls)
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
	repo, notifier, svc := newPublishTestService()
	repo.newCommentResult = repoResult

	created, err := svc.PublishNewComment(t.Context(), &NewCommentParams{
		PostID:   repoResult.PostID,
		AuthorID: repoResult.AuthorID,
		Body:     "исходный",
	})

	require.NoError(t, err)
	require.Same(t, repoResult, created)
	require.Equal(t, 1, notifier.notifyCreatedCalls)
	require.Same(t, repoResult, notifier.notifyCreatedInput)
	require.NotEqual(t, repo.newCommentInput.ID, created.ID)
}

func TestService_PublishNewComment_RepositoryFails(t *testing.T) {
	t.Parallel()

	repo, notifier, service := newPublishTestService()
	repo.newCommentErr = errRepo

	created, err := service.PublishNewComment(t.Context(), &NewCommentParams{
		PostID:   uuid.New(),
		AuthorID: uuid.New(),
		Body:     "комментарий",
	})

	require.Nil(t, created)
	require.ErrorContains(t, err, "publish comment")
	require.ErrorIs(t, err, repo.newCommentErr)
	require.Equal(t, 1, repo.newCommentCalls)
	require.Zero(t, notifier.notifyCreatedCalls)
}

func TestService_PublishNewComment_IgnoresNotifierFail(t *testing.T) {
	t.Parallel()

	repo, notifier, svc := newPublishTestService()
	notifier.err = errNotifier

	created, err := svc.PublishNewComment(t.Context(), &NewCommentParams{
		PostID:   uuid.New(),
		AuthorID: uuid.New(),
		Body:     "комментарий",
	})

	require.NoError(t, err)
	require.Same(t, repo.newCommentInput, created)
	require.Equal(t, 1, notifier.notifyCreatedCalls)
}

func TestService_PublishNewComment_IsSavedBeforeNotifying(t *testing.T) {
	t.Parallel()

	var calls []string
	repo, notifier, service := newPublishTestService()
	repo.onNewComment = func(*domain.Comment) {
		calls = append(calls, "saved")
	}
	notifier.onNotifyCreated = func(*domain.Comment) {
		calls = append(calls, "notified")
	}

	_, err := service.PublishNewComment(t.Context(), &NewCommentParams{
		PostID:   uuid.New(),
		AuthorID: uuid.New(),
		Body:     "комментарий",
	})

	require.NoError(t, err)
	require.Equal(t, []string{"saved", "notified"}, calls)
}
