package comment

import (
	"strconv"
	"testing"
	"time"

	"github.com/audworth/comments-system/internal/application"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestService_List(t *testing.T) {
	t.Parallel()

	postID, parentID := uuid.New(), uuid.New()
	after := &CommentPosition{CreatedAt: time.Now().UTC(), ID: uuid.New()}
	params := &ListParams{
		PostID:   postID,
		ParentID: &parentID,
		Limit:    25,
		After:    after,
	}
	want := &Page{
		Comments: []domain.Comment{{ID: uuid.New(), PostID: postID, ParentID: &parentID}},
		EndCursor: &CommentPosition{
			CreatedAt: time.Now().UTC(),
			ID:        uuid.New(),
		},
		HasNextPage: true,
	}

	repo, _, svc := newPublishTestService()
	repo.listChildrenResult = want

	got, err := svc.List(t.Context(), params)

	require.NoError(t, err)
	require.Same(t, repo.listChildrenResult, got)
	require.Equal(t, 1, repo.listChildrenCalls)
	require.Equal(t, params, repo.listChildrenInput)
}

func TestService_List_AcceptsBoundaryLimits(t *testing.T) {
	t.Parallel()

	for _, limit := range []int{1, 100} {
		t.Run(strconv.Itoa(limit), func(t *testing.T) {
			t.Parallel()

			repo, _, svc := newPublishTestService()
			repo.listChildrenResult = &Page{}

			got, err := svc.List(t.Context(), &ListParams{PostID: uuid.New(), Limit: limit})
			require.NoError(t, err)
			require.Same(t, repo.listChildrenResult, got)
			require.Equal(t, 1, repo.listChildrenCalls)
			require.Equal(t, limit, repo.listChildrenInput.Limit)
		})
	}
}

func TestService_List_RejectsInvalidLimit(t *testing.T) {
	t.Parallel()

	for _, limit := range []int{-1, 0, 101} {
		t.Run(strconv.Itoa(limit), func(t *testing.T) {
			t.Parallel()

			repo, _, svc := newPublishTestService()
			page, err := svc.List(t.Context(), &ListParams{PostID: uuid.New(), Limit: limit})

			require.Nil(t, page)
			require.ErrorIs(t, err, application.ErrInvalidPageSize)
			require.Zero(t, repo.listChildrenCalls)
		})
	}
}

func TestService_List_RepositoryFail(t *testing.T) {
	t.Parallel()

	postID := uuid.New()
	repo, _, svc := newPublishTestService()
	repo.listChildrenErr = errRepo

	page, err := svc.List(t.Context(), &ListParams{PostID: postID, Limit: 10})

	require.Nil(t, page)
	require.ErrorContains(t, err, "list comments for post "+postID.String())
	require.ErrorIs(t, err, repo.listChildrenErr)
	require.Equal(t, 1, repo.listChildrenCalls)
}
