package post

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

	after := &Position{CreatedAt: time.Now().UTC(), ID: uuid.New()}
	params := ListParams{Limit: 25, After: after}
	want := &Page{
		Items:       []domain.Post{{ID: uuid.New()}},
		Next:        &Position{CreatedAt: time.Now().UTC(), ID: uuid.New()},
		HasNextPage: true,
	}
	repo, svc := newTestService()
	repo.listResult = want

	got, err := svc.List(t.Context(), params)

	require.NoError(t, err)
	require.Same(t, want, got)
	require.Equal(t, 1, repo.listCalls)
	require.Equal(t, params, repo.listInput)
}

func TestService_List_AcceptsBoundaryLimits(t *testing.T) {
	t.Parallel()

	for _, limit := range []int{1, 100} {
		t.Run(strconv.Itoa(limit), func(t *testing.T) {
			t.Parallel()

			repo, svc := newTestService()
			repo.listResult = &Page{}

			got, err := svc.List(t.Context(), ListParams{Limit: limit})

			require.NoError(t, err)
			require.Same(t, repo.listResult, got)
			require.Equal(t, 1, repo.listCalls)
			require.Equal(t, limit, repo.listInput.Limit)
		})
	}
}

func TestService_List_RejectsInvalidLimit(t *testing.T) {
	t.Parallel()

	for _, limit := range []int{-100000, 0, 1010000000} {
		t.Run(strconv.Itoa(limit), func(t *testing.T) {
			t.Parallel()

			repo, svc := newTestService()
			page, err := svc.List(t.Context(), ListParams{Limit: limit})

			require.Nil(t, page)
			require.ErrorIs(t, err, application.ErrInvalidPageSize)
			require.Zero(t, repo.listCalls)
		})
	}
}

func TestService_List_RepositoryFail(t *testing.T) {
	t.Parallel()

	repo, svc := newTestService()
	repo.listErr = ErrNotFound

	page, err := svc.List(t.Context(), ListParams{Limit: 10})

	require.Nil(t, page)
	require.ErrorContains(t, err, "list posts")
	require.ErrorIs(t, err, repo.listErr)
	require.Equal(t, 1, repo.listCalls)
}
