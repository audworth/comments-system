package post

import (
	"log/slog"
	"testing"

	"go.uber.org/mock/gomock"
)

func newTestService(t *testing.T) (*MockRepository, *Service) {
	t.Helper()

	repo := NewMockRepository(gomock.NewController(t))
	return repo, NewService(repo, slog.New(slog.DiscardHandler))
}
