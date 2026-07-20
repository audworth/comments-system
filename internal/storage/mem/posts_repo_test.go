package mem

import (
	"log/slog"
	"testing"

	"github.com/audworth/comments-system/internal/application/post"
	"github.com/audworth/comments-system/internal/application/user"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestPostRepository_PublishAndGetByID(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)
	u := domain.User{ID: testID(1), Name: "user_1"}
	want := domain.Post{
		ID:              testID(10),
		AuthorID:        u.ID,
		Title:           "post",
		Body:            "body",
		CommentsEnabled: true,
		CreatedAt:       testTime,
		UpdatedAt:       testTime,
	}
	repo := NewPostRepository(newTestDB(t, []domain.User{u}, nil, nil), logger)

	created, err := repo.Publish(t.Context(), &want)
	require.NoError(t, err)
	require.Equal(t, &want, created)

	got, err := repo.GetByID(t.Context(), want.ID)
	require.NoError(t, err)
	require.Equal(t, &want, got)
}

func TestPostRepository_Publish_RejectsUnknownAuthor(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)
	repo := NewPostRepository(newTestDB(t, nil, nil, nil), logger)
	p := domain.Post{ID: testID(10), AuthorID: testID(1)}

	created, err := repo.Publish(t.Context(), &p)

	require.Nil(t, created)
	require.ErrorIs(t, err, user.ErrNotFound)
}

func TestPostRepository_Publish_RejectsDuplicateID(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)
	u := domain.User{ID: testID(1), Name: "user_1"}
	p := domain.Post{ID: testID(10), AuthorID: u.ID, CreatedAt: testTime}
	repo := NewPostRepository(newTestDB(t, []domain.User{u}, []domain.Post{p}, nil), logger)

	created, err := repo.Publish(t.Context(), &p)

	require.Nil(t, created)
	require.ErrorContains(t, err, "already exists")
}

func TestPostRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)
	repo := NewPostRepository(newTestDB(t, nil, nil, nil), logger)

	got, err := repo.GetByID(t.Context(), testID(10))

	require.Nil(t, got)
	require.ErrorIs(t, err, post.ErrNotFound)
}

func TestPostRepository_List_PaginatesByCreatedAtAndID(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)
	u := domain.User{ID: testID(1), Name: "user_1"}
	oldest := domain.Post{ID: testID(10), AuthorID: u.ID, CreatedAt: testTime.Add(-1)}
	middle := domain.Post{ID: testID(11), AuthorID: u.ID, CreatedAt: testTime}
	newest := domain.Post{ID: testID(12), AuthorID: u.ID, CreatedAt: testTime}
	repo := NewPostRepository(
		newTestDB(t, []domain.User{u}, []domain.Post{middle, oldest, newest}, nil),
		logger,
	)

	first, err := repo.List(t.Context(), post.ListParams{Limit: 2})
	require.NoError(t, err)
	require.Equal(t, []domain.Post{newest, middle}, first.Posts)
	require.True(t, first.HasNextPage)
	require.Equal(t, &post.Position{CreatedAt: middle.CreatedAt, ID: middle.ID}, first.EndCursor)

	second, err := repo.List(t.Context(), post.ListParams{Limit: 2, After: first.EndCursor})
	require.NoError(t, err)
	require.Equal(t, []domain.Post{oldest}, second.Posts)
	require.False(t, second.HasNextPage)
	require.Equal(t, &post.Position{CreatedAt: oldest.CreatedAt, ID: oldest.ID}, second.EndCursor)
}

func TestPostRepository_SetCommentsEnabled(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)
	u := domain.User{ID: testID(1), Name: "user_1"}
	p := domain.Post{
		ID:              testID(10),
		AuthorID:        u.ID,
		CommentsEnabled: true,
		CreatedAt:       testTime,
		UpdatedAt:       testTime,
	}
	repo := NewPostRepository(newTestDB(t, []domain.User{u}, []domain.Post{p}, nil), logger)

	updated, err := repo.SetCommentsEnabled(t.Context(), p.ID, u.ID, false)

	require.NoError(t, err)
	require.False(t, updated.CommentsEnabled)
	require.NotEqual(t, p.UpdatedAt, updated.UpdatedAt)

	stored, err := repo.GetByID(t.Context(), p.ID)
	require.NoError(t, err)
	require.Equal(t, updated, stored)
}

func TestPostRepository_SetCommentsEnabled_RejectsForbiddenAndMissingPost(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)
	u := domain.User{ID: testID(1), Name: "user_1"}
	p := domain.Post{ID: testID(10), AuthorID: u.ID}
	repo := NewPostRepository(newTestDB(t, []domain.User{u}, []domain.Post{p}, nil), logger)

	updated, err := repo.SetCommentsEnabled(t.Context(), p.ID, testID(2), false)
	require.Nil(t, updated)
	require.ErrorIs(t, err, post.ErrForbidden)

	updated, err = repo.SetCommentsEnabled(t.Context(), testID(99), u.ID, false)
	require.Nil(t, updated)
	require.ErrorIs(t, err, post.ErrNotFound)
}
