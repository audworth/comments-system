package comment

import (
	"testing"
	"time"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestService_CommentByID(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	expected := &domain.Comment{
		ID:        id,
		PostID:    uuid.New(),
		AuthorID:  uuid.New(),
		Body:      "комментарий",
		CreatedAt: time.Now().UTC(),
	}

	repo, _, svc := newPublishTestService()
	repo.commentByIDResult = expected

	actual, err := svc.CommentByID(t.Context(), expected.ID)

	require.NoError(t, err)
	require.Same(t, expected, actual)
	require.Equal(t, 1, repo.commentByIDCalls)
	require.Equal(t, expected.ID, repo.commentByIDInput)
}

func TestService_CommentByID_RepositoryFail(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	repo, _, svc := newPublishTestService()
	repo.commentByIDErr = errRepo

	comment, err := svc.CommentByID(t.Context(), id)

	require.Nil(t, comment)
	require.ErrorContains(t, err, "get comment "+id.String())
	require.ErrorIs(t, err, repo.commentByIDErr)
	require.Equal(t, 1, repo.commentByIDCalls)
	require.Equal(t, id, repo.commentByIDInput)
}
