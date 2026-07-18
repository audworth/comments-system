package post

import (
	"strings"
	"testing"
	"time"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestService_PublishNewPost(t *testing.T) {
	t.Parallel()

	for _, commentsEnabled := range []bool{false, true} {
		commentsEnabled := commentsEnabled
		t.Run(map[bool]string{false: "disabled", true: "enabled"}[commentsEnabled], func(t *testing.T) {
			t.Parallel()

			repo, svc := newTestService()
			authorID := uuid.New()
			before := time.Now().UTC()

			created, err := svc.PublishNewPost(t.Context(), NewPostParams{
				AuthorID:        authorID,
				Title:           "  title\n",
				Body:            "\tbody\r\n",
				CommentsEnabled: commentsEnabled,
			})
			after := time.Now().UTC()

			require.NoError(t, err)
			require.NotNil(t, created)
			require.NotEqual(t, uuid.Nil, created.ID)
			require.Equal(t, authorID, created.AuthorID)
			require.Equal(t, "title", created.Title)
			require.Equal(t, "body", created.Body)
			require.Equal(t, commentsEnabled, created.CommentsEnabled)
			require.WithinRange(t, created.CreatedAt, before, after)
			require.Equal(t, time.UTC, created.CreatedAt.Location())
			require.Equal(t, 1, repo.newPostCalls)
			require.Same(t, created, repo.newPostInput)
		})
	}
}

func TestService_PublishNewPost_AcceptsLongUnicodeContent(t *testing.T) {
	t.Parallel()

	repo, svc := newTestService()
	title := strings.Repeat("я", 10_000)
	body := strings.Repeat("💨", 10_000)

	created, err := svc.PublishNewPost(t.Context(), NewPostParams{
		AuthorID: uuid.New(),
		Title:    title,
		Body:     body,
	})

	require.NoError(t, err)
	require.Equal(t, title, created.Title)
	require.Equal(t, body, created.Body)
	require.Equal(t, 1, repo.newPostCalls)
}

func TestService_PublishNewPost_RejectsInvalidContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		title string
		body  string
		want  error
	}{
		{name: "пустой заголовок", title: "", body: "body", want: domain.ErrEmptyPostTitle},
		{name: "whitespace заголовок", title: " \t\r\n", body: "body", want: domain.ErrEmptyPostTitle},
		{name: "пустое тело", title: "title", body: "", want: domain.ErrEmptyPostBody},
		{name: "whitespace тело", title: "title", body: " \t\r\n", want: domain.ErrEmptyPostBody},
		{name: "сначала валидируется заголовок", title: "", body: "", want: domain.ErrEmptyPostTitle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo, svc := newTestService()
			created, err := svc.PublishNewPost(t.Context(), NewPostParams{
				AuthorID: uuid.New(),
				Title:    tt.title,
				Body:     tt.body,
			})

			require.Nil(t, created)
			require.ErrorContains(t, err, "invalid post")
			require.ErrorIs(t, err, tt.want)
			require.Zero(t, repo.newPostCalls)
		})
	}
}

func TestService_PublishNewPost_ReturnsRepositoryResult(t *testing.T) {
	t.Parallel()

	repositoryResult := &domain.Post{
		ID:              uuid.New(),
		AuthorID:        uuid.New(),
		Title:           "заголовок",
		Body:            "тело",
		CommentsEnabled: true,
		CreatedAt:       time.Now().UTC().Add(time.Second),
	}
	repo, svc := newTestService()
	repo.newPostResult = repositoryResult

	created, err := svc.PublishNewPost(t.Context(), NewPostParams{
		AuthorID: repositoryResult.AuthorID,
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
