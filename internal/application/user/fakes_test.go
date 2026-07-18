package user

import (
	"context"
	"errors"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

var (
	errRepo = errors.New("db unavailable")
)

type fakeRepo struct {
	userByIDResult *domain.User
	userByIDErr    error
	userByIDCalls  int
	userByIDInput  uuid.UUID
}

func (f *fakeRepo) UserByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	f.userByIDCalls++
	f.userByIDInput = id

	return f.userByIDResult, f.userByIDErr
}

func newTestService() (*fakeRepo, *Service) {
	repo := &fakeRepo{}

	return repo, NewService(repo)
}
