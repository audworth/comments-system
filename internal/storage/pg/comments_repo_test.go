package pg

import (
	"testing"
	"time"

	"github.com/audworth/comments-system/internal/application/comment"
	"github.com/audworth/comments-system/internal/application/user"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCommentsRepositoryIntegration_PublishAndGetReply(t *testing.T) {
	t.Parallel()

	db := newPGFixture(t)
	post := publishTestPost(t, db, testID(10), true, testTime)
	root := domain.Comment{
		ID:        testID(20),
		PostID:    post.ID,
		AuthorID:  db.user.ID,
		Body:      "root",
		CreatedAt: testTime,
	}
	createdRoot, err := db.comments.Publish(t.Context(), &root)
	require.NoError(t, err)
	require.Equal(t, root, *createdRoot)

	reply := domain.Comment{
		ID:        testID(21),
		PostID:    post.ID,
		ParentID:  &root.ID,
		AuthorID:  db.user.ID,
		Body:      "reply",
		CreatedAt: testTime.Add(time.Second),
	}
	createdReply, err := db.comments.Publish(t.Context(), &reply)
	require.NoError(t, err)
	require.Equal(t, reply, *createdReply)

	stored, err := db.comments.GetByID(t.Context(), reply.ID)
	require.NoError(t, err)
	require.Equal(t, reply, *stored)
}

func TestCommentsRepositoryIntegration_PublishRejectsInvalidReferences(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prepare func(t *testing.T, fixture *pgFixture, comm *domain.Comment)
		wantErr error
	}{
		{
			name: "пост не найден",
			prepare: func(_ *testing.T, _ *pgFixture, comm *domain.Comment) {
				comm.PostID = uuid.New()
			},
			wantErr: comment.ErrPostNotFound,
		},
		{
			name: "комментарии отключены",
			prepare: func(t *testing.T, fixture *pgFixture, comm *domain.Comment) {
				_, err := fixture.posts.SetCommentsEnabled(t.Context(), comm.PostID, fixture.user.ID, false)
				require.NoError(t, err)
			},
			wantErr: comment.ErrCommentsDisabled,
		},
		{
			name: "родитель не найден",
			prepare: func(_ *testing.T, _ *pgFixture, comm *domain.Comment) {
				parentID := uuid.New()
				comm.ParentID = &parentID
			},
			wantErr: comment.ErrParentNotFound,
		},
		{
			name: "родитель относится к другому посту",
			prepare: func(t *testing.T, fixture *pgFixture, comm *domain.Comment) {
				otherPost := publishTestPost(t, fixture, testID(11), true, testTime)
				parent := domain.Comment{
					ID:        testID(22),
					PostID:    otherPost.ID,
					AuthorID:  fixture.user.ID,
					Body:      "other root",
					CreatedAt: testTime,
				}
				_, err := fixture.comments.Publish(t.Context(), &parent)
				require.NoError(t, err)
				comm.ParentID = &parent.ID
			},
			wantErr: comment.ErrParentNotFound,
		},
		{
			name: "автор не найден",
			prepare: func(_ *testing.T, _ *pgFixture, comm *domain.Comment) {
				comm.AuthorID = uuid.New()
			},
			wantErr: user.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := newPGFixture(t)
			post := publishTestPost(t, db, testID(10), true, testTime)
			comm := domain.Comment{
				ID:        testID(20),
				PostID:    post.ID,
				AuthorID:  db.user.ID,
				Body:      "comment",
				CreatedAt: testTime,
			}
			tt.prepare(t, db, &comm)

			created, err := db.comments.Publish(t.Context(), &comm)
			require.Nil(t, created)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestCommentsRepositoryIntegration_ListAndListBatch(t *testing.T) {
	t.Parallel()

	db := newPGFixture(t)
	post := publishTestPost(t, db, testID(10), true, testTime)
	oldest := domain.Comment{ID: testID(20), PostID: post.ID, AuthorID: db.user.ID, Body: "oldest", CreatedAt: testTime.Add(-time.Second)}
	middle := domain.Comment{ID: testID(21), PostID: post.ID, AuthorID: db.user.ID, Body: "middle", CreatedAt: testTime}
	newest := domain.Comment{ID: testID(22), PostID: post.ID, AuthorID: db.user.ID, Body: "newest", CreatedAt: testTime}
	for _, comm := range []domain.Comment{middle, oldest, newest} {
		_, err := db.comments.Publish(t.Context(), &comm)
		require.NoError(t, err)
	}
	reply := domain.Comment{
		ID:        testID(23),
		PostID:    post.ID,
		ParentID:  &oldest.ID,
		AuthorID:  db.user.ID,
		Body:      "reply",
		CreatedAt: testTime.Add(time.Second),
	}
	_, err := db.comments.Publish(t.Context(), &reply)
	require.NoError(t, err)

	first, err := db.comments.List(t.Context(), comment.ListParams{PostID: post.ID, Limit: 2})
	require.NoError(t, err)
	require.Equal(t, []domain.Comment{newest, middle}, first.Comments)
	require.True(t, first.HasNextPage)
	require.Equal(t, &comment.Position{CreatedAt: middle.CreatedAt, ID: middle.ID}, first.EndCursor)

	second, err := db.comments.List(t.Context(), comment.ListParams{
		PostID: post.ID,
		Limit:  2,
		After:  first.EndCursor,
	})
	require.NoError(t, err)
	require.Equal(t, []domain.Comment{oldest}, second.Comments)
	require.False(t, second.HasNextPage)

	pages, err := db.comments.ListBatch(t.Context(), []comment.ListParams{
		{PostID: post.ID, Limit: 10},
		{PostID: post.ID, ParentID: &oldest.ID, Limit: 10},
	})
	require.NoError(t, err)
	require.Len(t, pages, 2)
	require.Equal(t, []domain.Comment{newest, middle, oldest}, pages[0].Comments)
	require.Equal(t, []domain.Comment{reply}, pages[1].Comments)
}
