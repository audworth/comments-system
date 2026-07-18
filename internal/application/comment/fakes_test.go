package comment

import (
	"context"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

type fakeRepo struct {
	newCommentResult *domain.Comment
	newCommentErr    error
	newCommentCalls  int
	newCommentInput  *domain.Comment
	onNewComment     func(*domain.Comment)
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

func (f *fakeRepo) CommentByID(context.Context, uuid.UUID) (*domain.Comment, error) {
	panic("TODO")
}

func (f *fakeRepo) ListChildren(context.Context, ListParams) (*Page, error) {
	panic("TODO")
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
