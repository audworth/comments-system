package dataloader

import (
	"context"

	"github.com/audworth/comments-system/internal/domain"
	grapherror "github.com/audworth/comments-system/internal/transport/graph/error"
	"github.com/google/uuid"
	"github.com/vikstrous/dataloadgen"
)

type userReader interface {
	UsersByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domain.User, error)
}

func GetUser(ctx context.Context, id string) (*domain.User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, grapherror.InvalidID("userId", err)
	}

	loaders, err := fromContext(ctx)
	if err != nil {
		return nil, err
	}

	return loaders.UserLoader.Load(ctx, userID)
}

func GetUsers(ctx context.Context, ids []string) ([]*domain.User, error) {
	userIDs := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return nil, grapherror.InvalidID("userIds", err)
		}
		userIDs[i] = parsed
	}

	loaders, err := fromContext(ctx)
	if err != nil {
		return nil, err
	}

	return loaders.UserLoader.LoadAll(ctx, userIDs)
}

func newUserLoader(users userReader) *dataloadgen.Loader[uuid.UUID, *domain.User] {
	return dataloadgen.NewMappedLoader(
		users.UsersByIDs,
		dataloadgen.WithWait(loaderWait),
		dataloadgen.WithBatchCapacity(loaderBatchCapacity),
	)
}
