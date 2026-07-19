package post

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

	after := &Position{CreatedAt: time.Now().UTC(), ID: uuid.New()}
	params := ListParams{Limit: 25, After: after}
	want := &Page{
		Posts:       []domain.Post{{ID: uuid.New()}},
		Next:        &Position{CreatedAt: time.Now().UTC(), ID: uuid.New()},
		HasNextPage: true,
	}
	repo, svc := newTestService(t)
	repo.EXPECT().List(gomock.Any(), params).Return(want, nil)

	got, err := svc.List(t.Context(), params)

	require.NoError(t, err)
	require.Same(t, want, got)
}

func TestService_List_AcceptsBoundaryLimits(t *testing.T) {
	t.Parallel()

	for _, limit := range []int{1, 100} {
		t.Run(strconv.Itoa(limit), func(t *testing.T) {
			t.Parallel()

			repo, svc := newTestService(t)
			want := &Page{}
			params := ListParams{Limit: limit}
			repo.EXPECT().List(gomock.Any(), params).Return(want, nil)

			got, err := svc.List(t.Context(), params)

			require.NoError(t, err)
			require.Same(t, want, got)
		})
	}
}

func TestService_List_RejectsInvalidLimit(t *testing.T) {
	t.Parallel()

	for _, limit := range []int{-100000, 0, 1010000000} {
		t.Run(strconv.Itoa(limit), func(t *testing.T) {
			t.Parallel()

			_, svc := newTestService(t)
			page, err := svc.List(t.Context(), ListParams{Limit: limit})

			require.Nil(t, page)
			require.ErrorIs(t, err, application.ErrInvalidPageSize)
		})
	}
}

func TestService_List_RepositoryFail(t *testing.T) {
	t.Parallel()

	repo, svc := newTestService(t)
	params := ListParams{Limit: 10}
	repo.EXPECT().List(gomock.Any(), params).Return(nil, ErrNotFound)

	page, err := svc.List(t.Context(), params)

	require.Nil(t, page)
	require.ErrorContains(t, err, "list posts")
	require.ErrorIs(t, err, ErrNotFound)
}
