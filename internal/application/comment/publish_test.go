package comment

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newPublishTestService() (*fakeRepo, *notifierSpy, *Service) {
	repo := &fakeRepo{}
	notifier := &notifierSpy{}

	return repo, notifier, NewService(repo, notifier)
}

func TestService_PublishNewComment(t *testing.T) {
	t.Parallel()

	repo, notif, svc := newPublishTestService()
	postID, authorID := uuid.New(), uuid.New()
	before := time.Now().UTC()

	created, err := svc.PublishNewComment(t.Context(), &NewCommentParams{
		PostID:   postID,
		AuthorID: authorID,
		Body:     "комментарий",
	})
	after := time.Now().UTC()

	require.NoError(t, err)
	require.NotNil(t, created)
	require.NotEqual(t, uuid.Nil, created.ID)
	require.Equal(t, postID, created.PostID)
	require.Nil(t, created.ParentID)
	require.Equal(t, authorID, created.AuthorID)
	require.Equal(t, "комментарий", created.Body)
	require.WithinRange(t, created.CreatedAt, before, after)

	require.Equal(t, 1, repo.newCommentCalls)
	require.Same(t, created, repo.newCommentInput)
	require.Equal(t, 1, notif.notifyCreatedCalls)
	require.Same(t, created, notif.notifyCreatedInput)
}

func TestService_PublishNewComment_WithParent(t *testing.T) {
	t.Parallel()

	repo, notif, svc := newPublishTestService()
	parentID := uuid.New()

	created, err := svc.PublishNewComment(t.Context(), &NewCommentParams{
		PostID:   uuid.New(),
		ParentID: &parentID,
		AuthorID: uuid.New(),
		Body:     "ответ",
	})

	require.NoError(t, err)
	require.NotNil(t, created)
	require.NotNil(t, created.ParentID)
	require.NotNil(t, repo.newCommentInput.ParentID)
	require.NotNil(t, notif.notifyCreatedInput.ParentID)
	require.Equal(t, parentID, *created.ParentID)
	require.Equal(t, parentID, *repo.newCommentInput.ParentID)
	require.Equal(t, parentID, *notif.notifyCreatedInput.ParentID)
}

func TestService_PublishNewComment_ReturnsAndNotifiesRepositoryResult(t *testing.T) {
	t.Parallel()

	repositoryResult := &domain.Comment{
		ID:        uuid.New(),
		PostID:    uuid.New(),
		AuthorID:  uuid.New(),
		Body:      "сохраненный",
		CreatedAt: time.Now().UTC().Add(time.Second),
	}
	repo, notif, svc := newPublishTestService()
	repo.newCommentResult = repositoryResult

	created, err := svc.PublishNewComment(t.Context(), &NewCommentParams{
		PostID:   repositoryResult.PostID,
		AuthorID: repositoryResult.AuthorID,
		Body:     "исходный",
	})

	require.NoError(t, err)
	require.Same(t, repositoryResult, created)
	require.Equal(t, 1, notif.notifyCreatedCalls)
	require.Same(t, repositoryResult, notif.notifyCreatedInput)
	require.NotEqual(t, repo.newCommentInput.ID, created.ID)
}

func TestService_PublishNewComment_AcceptsValidBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		body     string
		expected string
	}{
		{name: "текс", body: "комментарий", expected: "комментарий"},
		{name: "whitespaces", body: "  комментарий\n\r\r\r\n\n\n\t\t\t", expected: "комментарий"},
		{name: "2000 ASCII", body: strings.Repeat("a", 2000), expected: strings.Repeat("a", 2000)},
		{name: "2000 UTF-8", body: strings.Repeat("я", 2000), expected: strings.Repeat("я", 2000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo, notif, svc := newPublishTestService()
			created, err := svc.PublishNewComment(t.Context(), &NewCommentParams{
				PostID:   uuid.New(),
				AuthorID: uuid.New(),
				Body:     tt.body,
			})

			require.NoError(t, err)
			require.Equal(t, tt.expected, created.Body)
			require.Equal(t, 1, repo.newCommentCalls)
			require.Equal(t, 1, notif.notifyCreatedCalls)
		})
	}
}

func TestService_PublishNewComment_RejectsInvalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		body        string
		expectedErr error
	}{
		{name: "пустой", body: "", expectedErr: domain.ErrEmptyComment},
		{name: "пробелы", body: "     ", expectedErr: domain.ErrEmptyComment},
		{name: "whitespaces", body: "\t\t\t\t\n\n\r\t\n\r ", expectedErr: domain.ErrEmptyComment},
		{name: "2001 ASCII", body: strings.Repeat("a", 2001), expectedErr: domain.ErrCommentTooLong},
		{name: "2001 UTF-8", body: strings.Repeat("я", 2001), expectedErr: domain.ErrCommentTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo, notif, svc := newPublishTestService()
			created, err := svc.PublishNewComment(t.Context(), &NewCommentParams{
				PostID:   uuid.New(),
				AuthorID: uuid.New(),
				Body:     tt.body,
			})

			require.Nil(t, created)
			require.ErrorContains(t, err, "invalid comment")
			require.ErrorIs(t, err, tt.expectedErr)
			require.Zero(t, repo.newCommentCalls)
			require.Zero(t, notif.notifyCreatedCalls)
		})
	}
}

func TestService_PublishNewComment_RepositoryFails(t *testing.T) {
	t.Parallel()

	repo, notifier, service := newPublishTestService()
	repo.newCommentErr = errors.New("db error")

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
	notifier.err = errors.New("unavailable")

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
