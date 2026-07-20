package mem

import (
	"log/slog"
	"testing"

	"github.com/audworth/comments-system/internal/application/comment"
	"github.com/audworth/comments-system/internal/application/user"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/audworth/comments-system/internal/platform/db/inmem"
	"github.com/stretchr/testify/require"
)

func TestCommentsRepository_PublishAndGetByID(t *testing.T) {
	t.Parallel()

	u := domain.User{ID: testID(1), Name: "user_1"}
	p := domain.Post{ID: testID(10), AuthorID: u.ID, CommentsEnabled: true}
	parent := domain.Comment{ID: testID(20), PostID: p.ID, AuthorID: u.ID, Body: "parent", CreatedAt: testTime}
	parentID := parent.ID
	want := domain.Comment{
		ID:        testID(21),
		PostID:    p.ID,
		ParentID:  &parentID,
		AuthorID:  u.ID,
		Body:      "reply",
		CreatedAt: testTime.Add(1),
	}
	logger := slog.New(slog.DiscardHandler)
	repo := NewCommentsRepository(newTestDB(t, []domain.User{u}, []domain.Post{p}, []domain.Comment{parent}), logger)

	created, err := repo.Publish(t.Context(), &want)
	require.NoError(t, err)
	require.Equal(t, &want, created)

	got, err := repo.GetByID(t.Context(), want.ID)
	require.NoError(t, err)
	require.Equal(t, &want, got)
}

func TestCommentsRepository_Publish_RejectsInvalidReferences(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prepare func(t *testing.T) (*CommentsRepository, domain.Comment)
		wantErr error
	}{
		{
			name: "пост не найден",
			prepare: func(t *testing.T) (*CommentsRepository, domain.Comment) {
				logger := slog.New(slog.DiscardHandler)
				u := domain.User{ID: testID(1), Name: "user_1"}
				comm := domain.Comment{ID: testID(20), PostID: testID(10), AuthorID: u.ID}
				return NewCommentsRepository(newTestDB(t, []domain.User{u}, nil, nil), logger), comm
			},
			wantErr: comment.ErrPostNotFound,
		},
		{
			name: "комментарии запрещены",
			prepare: func(t *testing.T) (*CommentsRepository, domain.Comment) {
				logger := slog.New(slog.DiscardHandler)
				u := domain.User{ID: testID(1), Name: "user_1"}
				p := domain.Post{ID: testID(10), AuthorID: u.ID, CommentsEnabled: false}
				comm := domain.Comment{ID: testID(20), PostID: p.ID, AuthorID: u.ID}
				return NewCommentsRepository(newTestDB(t, []domain.User{u}, []domain.Post{p}, nil), logger), comm
			},
			wantErr: comment.ErrCommentsDisabled,
		},
		{
			name: "родительский комментарий не найден",
			prepare: func(t *testing.T) (*CommentsRepository, domain.Comment) {
				logger := slog.New(slog.DiscardHandler)
				u := domain.User{ID: testID(1), Name: "user_1"}
				p := domain.Post{ID: testID(10), AuthorID: u.ID, CommentsEnabled: true}
				parentID := testID(99)
				comm := domain.Comment{ID: testID(20), PostID: p.ID, ParentID: &parentID, AuthorID: u.ID}
				return NewCommentsRepository(newTestDB(t, []domain.User{u}, []domain.Post{p}, nil), logger), comm
			},
			wantErr: comment.ErrParentNotFound,
		},
		{
			name: "родительский комментарий относится к другому посту",
			prepare: func(t *testing.T) (*CommentsRepository, domain.Comment) {
				logger := slog.New(slog.DiscardHandler)
				u := domain.User{ID: testID(1), Name: "user_1"}
				p1 := domain.Post{ID: testID(10), AuthorID: u.ID, CommentsEnabled: true}
				p2 := domain.Post{ID: testID(11), AuthorID: u.ID, CommentsEnabled: true}
				parent := domain.Comment{ID: testID(20), PostID: p2.ID, AuthorID: u.ID}
				parentID := parent.ID
				comm := domain.Comment{ID: testID(21), PostID: p1.ID, ParentID: &parentID, AuthorID: u.ID}
				return NewCommentsRepository(newTestDB(
					t,
					[]domain.User{u},
					[]domain.Post{p1, p2},
					[]domain.Comment{parent},
				), logger), comm
			},
			wantErr: comment.ErrParentNotFound,
		},
		{
			name: "автор не найден",
			prepare: func(t *testing.T) (*CommentsRepository, domain.Comment) {
				logger := slog.New(slog.DiscardHandler)
				u := domain.User{ID: testID(1), Name: "user_1"}
				p := domain.Post{ID: testID(10), AuthorID: u.ID, CommentsEnabled: true}
				comm := domain.Comment{ID: testID(20), PostID: p.ID, AuthorID: testID(2)}
				return NewCommentsRepository(newTestDB(t, []domain.User{u}, []domain.Post{p}, nil), logger), comm
			},
			wantErr: user.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo, comm := tt.prepare(t)
			created, err := repo.Publish(t.Context(), &comm)

			require.Nil(t, created)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCommentsRepository_Publish_RejectsDuplicateID(t *testing.T) {
	t.Parallel()

	u := domain.User{ID: testID(1), Name: "user_1"}
	p := domain.Post{ID: testID(10), AuthorID: u.ID, CommentsEnabled: true}
	comm := domain.Comment{ID: testID(20), PostID: p.ID, AuthorID: u.ID}
	logger := slog.New(slog.DiscardHandler)
	repo := NewCommentsRepository(newTestDB(t, []domain.User{u}, []domain.Post{p}, []domain.Comment{comm}), logger)

	created, err := repo.Publish(t.Context(), &comm)

	require.Nil(t, created)
	require.ErrorContains(t, err, "already exists")
}

func TestCommentsRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)
	repo := NewCommentsRepository(inmem.New(), logger)

	got, err := repo.GetByID(t.Context(), testID(20))

	require.Nil(t, got)
	require.ErrorIs(t, err, comment.ErrNotFound)
}

func TestCommentsRepository_List_PaginatesRootsAndReplies(t *testing.T) {
	t.Parallel()

	u := domain.User{ID: testID(1), Name: "user_1"}
	p := domain.Post{ID: testID(10), AuthorID: u.ID, CommentsEnabled: true}
	oldest := domain.Comment{ID: testID(20), PostID: p.ID, AuthorID: u.ID, CreatedAt: testTime.Add(-1)}
	middle := domain.Comment{ID: testID(21), PostID: p.ID, AuthorID: u.ID, CreatedAt: testTime}
	newest := domain.Comment{ID: testID(22), PostID: p.ID, AuthorID: u.ID, CreatedAt: testTime}
	parentID := oldest.ID
	reply := domain.Comment{ID: testID(23), PostID: p.ID, ParentID: &parentID, AuthorID: u.ID, CreatedAt: testTime.Add(1)}
	logger := slog.New(slog.DiscardHandler)
	repo := NewCommentsRepository(newTestDB(
		t,
		[]domain.User{u},
		[]domain.Post{p},
		[]domain.Comment{middle, reply, oldest, newest},
	), logger)

	first, err := repo.List(t.Context(), comment.ListParams{PostID: p.ID, Limit: 2})
	require.NoError(t, err)
	require.Equal(t, []domain.Comment{newest, middle}, first.Comments)
	require.True(t, first.HasNextPage)
	require.Equal(t, &comment.Position{CreatedAt: middle.CreatedAt, ID: middle.ID}, first.EndCursor)

	second, err := repo.List(t.Context(), comment.ListParams{
		PostID: p.ID,
		Limit:  2,
		After:  first.EndCursor,
	})
	require.NoError(t, err)
	require.Equal(t, []domain.Comment{oldest}, second.Comments)
	require.False(t, second.HasNextPage)

	replies, err := repo.List(t.Context(), comment.ListParams{
		PostID:   p.ID,
		ParentID: &parentID,
		Limit:    10,
	})
	require.NoError(t, err)
	require.Equal(t, []domain.Comment{reply}, replies.Comments)
}

func TestCommentsRepository_ListBatch(t *testing.T) {
	t.Parallel()

	u := domain.User{ID: testID(1), Name: "user_1"}
	p := domain.Post{ID: testID(10), AuthorID: u.ID, CommentsEnabled: true}
	root := domain.Comment{ID: testID(20), PostID: p.ID, AuthorID: u.ID, CreatedAt: testTime}
	parentID := root.ID
	reply := domain.Comment{ID: testID(21), PostID: p.ID, ParentID: &parentID, AuthorID: u.ID, CreatedAt: testTime.Add(1)}
	logger := slog.New(slog.DiscardHandler)
	repo := NewCommentsRepository(newTestDB(
		t,
		[]domain.User{u},
		[]domain.Post{p},
		[]domain.Comment{root, reply},
	), logger)

	pages, err := repo.ListBatch(t.Context(), []comment.ListParams{
		{PostID: p.ID, Limit: 10},
		{PostID: p.ID, ParentID: &parentID, Limit: 10},
	})

	require.NoError(t, err)
	require.Len(t, pages, 2)
	require.Equal(t, []domain.Comment{root}, pages[0].Comments)
	require.Equal(t, []domain.Comment{reply}, pages[1].Comments)
}
