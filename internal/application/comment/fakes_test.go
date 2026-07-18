package comment

import (
	"context"
	"errors"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

var (
	errRepo     = errors.New("db unavailable")
	errNotifier = errors.New("notifier unavailable")
)

type fakeRepo struct {
	newCommentResult *domain.Comment
	newCommentErr    error
	newCommentCalls  int
	newCommentInput  *domain.Comment
	onNewComment     func(*domain.Comment)

	commentByIDResult *domain.Comment
	commentByIDErr    error
	commentByIDCalls  int
	commentByIDInput  uuid.UUID

	listChildrenResult *Page
	listChildrenErr    error
	listChildrenCalls  int
	listChildrenInput  *ListParams
}

func (f *fakeRepo) NewComment(_ context.Context, comment *domain.Comment) (*domain.Comment, error) {
	f.newCommentCalls++
	f.newCommentInput = comment

	if f.onNewComment != nil {
		f.onNewComment(comment)
	}
	if f.newCommentErr != nil {
		return nil, f.newCommentErr
	}
	if f.newCommentResult != nil {
		return f.newCommentResult, nil
	}

	return comment, nil
}

func (f *fakeRepo) CommentByID(_ context.Context, id uuid.UUID) (*domain.Comment, error) {
	f.commentByIDCalls++
	f.commentByIDInput = id

	return f.commentByIDResult, f.commentByIDErr
}

func (f *fakeRepo) ListChildren(_ context.Context, params *ListParams) (*Page, error) {
	f.listChildrenCalls++
	f.listChildrenInput = params

	return f.listChildrenResult, f.listChildrenErr
}

type notifierSpy struct {
	err                error
	notifyCreatedCalls int
	notifyCreatedInput *domain.Comment
	onNotifyCreated    func(*domain.Comment)
}

func (s *notifierSpy) NotifyCreated(_ context.Context, comment *domain.Comment) error {
	s.notifyCreatedCalls++
	s.notifyCreatedInput = comment

	if s.onNotifyCreated != nil {
		s.onNotifyCreated(comment)
	}

	return s.err
}

func newPublishTestService() (*fakeRepo, *notifierSpy, *Service) {
	repo := &fakeRepo{}
	notifier := &notifierSpy{}

	return repo, notifier, NewService(repo, notifier)
}
