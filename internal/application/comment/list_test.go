package comment

import (
	"strconv"
	"testing"

	"github.com/audworth/comments-system/internal/application"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_List_AcceptsBoundaryLimits(t *testing.T) {
	t.Parallel()

	for _, limit := range []int{1, 100} {
		t.Run(strconv.Itoa(limit), func(t *testing.T) {
			t.Parallel()

			repo, _, svc := newTestService(t)
			params := ListParams{PostID: uuid.New(), Limit: limit}
			want := &Page{}
			repo.EXPECT().List(gomock.Any(), params).Return(want, nil)

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
	repo.EXPECT().List(gomock.Any(), params).Return(nil, errRepo)

	page, err := svc.List(t.Context(), params)

	require.Nil(t, page)
	require.ErrorContains(t, err, "list comments for post "+postID.String())
	require.ErrorIs(t, err, errRepo)
}
