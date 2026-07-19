package comment

import (
	"strconv"
	"testing"
	"time"

	"github.com/audworth/comments-system/internal/application"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_List(t *testing.T) {
	t.Parallel()

	postID, parentID := uuid.New(), uuid.New()
	after := &Position{CreatedAt: time.Now().UTC(), ID: uuid.New()}
	params := ListParams{
		PostID:   postID,
		ParentID: &parentID,
		Limit:    25,
		After:    after,
	}
	want := &Page{
		Comments: []domain.Comment{{ID: uuid.New(), PostID: postID, ParentID: &parentID}},
		EndCursor: &Position{
			CreatedAt: time.Now().UTC(),
			ID:        uuid.New(),
		},
		HasNextPage: true,
	}

	repo, _, svc := newTestService(t)
	repo.EXPECT().ListChildren(gomock.Any(), params).Return(want, nil)

	got, err := svc.List(t.Context(), params)

	require.NoError(t, err)
	require.Same(t, want, got)
}

func TestService_List_AcceptsBoundaryLimits(t *testing.T) {
	t.Parallel()

	for _, limit := range []int{1, 100} {
		t.Run(strconv.Itoa(limit), func(t *testing.T) {
			t.Parallel()

			repo, _, svc := newTestService(t)
			params := ListParams{PostID: uuid.New(), Limit: limit}
			want := &Page{}
			repo.EXPECT().ListChildren(gomock.Any(), params).Return(want, nil)

			got, err := svc.List(t.Context(), params)
			require.NoError(t, err)
			require.Same(t, want, got)
		})
	}
}

func TestService_List_RejectsInvalidLimit(t *testing.T) {
	t.Parallel()

	for _, limit := range []int{-1, 0, 101} {
		t.Run(strconv.Itoa(limit), func(t *testing.T) {
			t.Parallel()

			_, _, svc := newTestService(t)
			page, err := svc.List(t.Context(), ListParams{PostID: uuid.New(), Limit: limit})

			require.Nil(t, page)
			require.ErrorIs(t, err, application.ErrInvalidPageSize)
		})
	}
}

func TestService_List_RepositoryFail(t *testing.T) {
	t.Parallel()

	postID := uuid.New()
	repo, _, svc := newTestService(t)
	params := ListParams{PostID: postID, Limit: 10}
	repo.EXPECT().ListChildren(gomock.Any(), params).Return(nil, errRepo)

	page, err := svc.List(t.Context(), params)

	require.Nil(t, page)
	require.ErrorContains(t, err, "list comments for post "+postID.String())
	require.ErrorIs(t, err, errRepo)
}
