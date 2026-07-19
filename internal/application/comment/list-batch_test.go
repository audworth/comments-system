package comment

import (
	"errors"
	"testing"

	"github.com/audworth/comments-system/internal/application"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_ListBatch(t *testing.T) {
	t.Parallel()

	params := []ListParams{
		{PostID: uuid.New(), Limit: 20},
		{PostID: uuid.New(), ParentID: uuidPointer(uuid.New()), Limit: 10},
	}
	want := []*Page{{}, {HasNextPage: true}}
	repo, _, svc := newTestService(t)
	repo.EXPECT().ListChildrenBatch(gomock.Any(), params).Return(want, nil)

	got, err := svc.ListBatch(t.Context(), params)

	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestService_ListBatch_RejectsInvalidLimit(t *testing.T) {
	t.Parallel()

	_, _, svc := newTestService(t)
	pages, err := svc.ListBatch(t.Context(), []ListParams{
		{PostID: uuid.New(), Limit: 20},
		{PostID: uuid.New(), Limit: 101},
	})

	require.Nil(t, pages)
	require.ErrorIs(t, err, application.ErrInvalidPageSize)
	require.ErrorContains(t, err, "comment page size 1")
}

func TestService_ListBatch_EmptyBatch(t *testing.T) {
	t.Parallel()

	_, _, svc := newTestService(t)
	pages, err := svc.ListBatch(t.Context(), nil)

	require.NoError(t, err)
	require.Empty(t, pages)
	require.NotNil(t, pages)
}

func TestService_ListBatch_RepositoryFails(t *testing.T) {
	t.Parallel()

	params := []ListParams{{PostID: uuid.New(), Limit: 20}}
	wantErr := errors.New("storage unavailable")
	repo, _, svc := newTestService(t)
	repo.EXPECT().ListChildrenBatch(gomock.Any(), params).Return(nil, wantErr)

	pages, err := svc.ListBatch(t.Context(), params)

	require.Nil(t, pages)
	require.ErrorIs(t, err, wantErr)
	require.ErrorContains(t, err, "list comment pages")
}

func TestService_ListBatch_RejectsWrongResultCount(t *testing.T) {
	t.Parallel()

	params := []ListParams{
		{PostID: uuid.New(), Limit: 20},
		{PostID: uuid.New(), Limit: 20},
	}
	repo, _, svc := newTestService(t)
	repo.EXPECT().ListChildrenBatch(gomock.Any(), params).Return([]*Page{{}}, nil)

	pages, err := svc.ListBatch(t.Context(), params)

	require.Nil(t, pages)
	require.ErrorContains(t, err, "returned 1 pages for 2 requests")
}

func uuidPointer(id uuid.UUID) *uuid.UUID {
	return &id
}
