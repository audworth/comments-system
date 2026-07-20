package comment

import (
	"errors"
	"log/slog"
	"testing"

	"go.uber.org/mock/gomock"
)

var (
	errRepo     = errors.New("db unavailable")
	errNotifier = errors.New("notifier unavailable")
)

func newTestService(t *testing.T) (*MockRepository, *MockNotifier, *Service) {
	t.Helper()

	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	notifier := NewMockNotifier(ctrl)
	subscriber := NewMockSubscriber(ctrl)
	logger := slog.New(slog.DiscardHandler)

	return repo, notifier, NewService(repo, notifier, subscriber, logger)
}
