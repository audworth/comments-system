package post

import (
	"testing"

	"go.uber.org/mock/gomock"
)

func newTestService(t *testing.T) (*MockRepository, *Service) {
	t.Helper()

	repo := NewMockRepository(gomock.NewController(t))
	return repo, NewService(repo)
}
