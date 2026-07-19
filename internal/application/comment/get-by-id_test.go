package comment

import (
	"testing"
	"time"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_GetByID(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	expected := &domain.Comment{
		ID:        id,
		PostID:    uuid.New(),
		AuthorID:  uuid.New(),
		Body:      "комментарий",
		CreatedAt: time.Now().UTC(),
	}

	repo, _, svc := newTestService(t)
	repo.EXPECT().GetByID(gomock.Any(), id).Return(expected, nil)

	actual, err := svc.GetByID(t.Context(), expected.ID)

	require.NoError(t, err)
	require.Same(t, expected, actual)
}

func TestService_GetByID_RepositoryFail(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	repo, _, svc := newTestService(t)
	repo.EXPECT().GetByID(gomock.Any(), id).Return(nil, errRepo)

	comment, err := svc.GetByID(t.Context(), id)

	require.Nil(t, comment)
	require.ErrorContains(t, err, "get comment "+id.String())
	require.ErrorIs(t, err, errRepo)
}
