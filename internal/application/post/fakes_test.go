package post

import (
	"context"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

type fakeRepo struct {
	newPostResult *domain.Post
	newPostErr    error
	newPostCalls  int
	newPostInput  *domain.Post

	postByIDResult *domain.Post
	postByIDErr    error
	postByIDCalls  int
	postByIDInput  uuid.UUID

	listResult *Page
	listErr    error
	listCalls  int
	listInput  ListParams

	setCommentsEnabledResult  *domain.Post
	setCommentsEnabledErr     error
	setCommentsEnabledCalls   int
	setCommentsEnabledPostID  uuid.UUID
	setCommentsEnabledAuthor  uuid.UUID
	setCommentsEnabledEnabled bool
}

func (f *fakeRepo) NewPost(_ context.Context, post *domain.Post) (*domain.Post, error) {
	f.newPostCalls++
	f.newPostInput = post

	if f.newPostErr != nil {
		return nil, f.newPostErr
	}
	if f.newPostResult != nil {
		return f.newPostResult, nil
	}

	return post, nil
}

func (f *fakeRepo) PostByID(_ context.Context, id uuid.UUID) (*domain.Post, error) {
	f.postByIDCalls++
	f.postByIDInput = id

	return f.postByIDResult, f.postByIDErr
}

func (f *fakeRepo) List(_ context.Context, params ListParams) (*Page, error) {
	f.listCalls++
	f.listInput = params

	return f.listResult, f.listErr
}

func (f *fakeRepo) SetCommentsEnabled(
	_ context.Context,
	postID uuid.UUID,
	author uuid.UUID,
	enabled bool,
) (*domain.Post, error) {
	f.setCommentsEnabledCalls++
	f.setCommentsEnabledPostID = postID
	f.setCommentsEnabledAuthor = author
	f.setCommentsEnabledEnabled = enabled

	return f.setCommentsEnabledResult, f.setCommentsEnabledErr
}

func newTestService() (*fakeRepo, *Service) {
	repo := &fakeRepo{}
	return repo, NewService(repo)
}
