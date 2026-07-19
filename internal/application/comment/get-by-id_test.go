package comment

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

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
