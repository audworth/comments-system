package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmptyPostTitle = errors.New("post title must not be empty")
	ErrEmptyPostBody  = errors.New("post body must not be empty")
	ErrNotPostAuthor  = errors.New("user is not the post author")
)

type Post struct {
	ID              uuid.UUID
	Title           string
	Body            string
	CommentsEnabled bool
	Author          User
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (p *Post) SetCommentEnabled(userID uuid.UUID, enabled bool) error {
	if p.Author.ID != userID {
		return ErrNotPostAuthor
	}

	p.CommentsEnabled = enabled

	return nil
}

func NewPost(
	id uuid.UUID,
	author User,
	title string,
	body string,
	commentsEnabled bool,
	createdAt time.Time,
) (*Post, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return &Post{}, ErrEmptyPostTitle
	}

	body = strings.TrimSpace(body)
	if body == "" {
		return &Post{}, ErrEmptyPostBody
	}

	return &Post{
		ID:              id,
		Author:          author,
		Title:           title,
		Body:            body,
		CommentsEnabled: commentsEnabled,
		CreatedAt:       createdAt,
		UpdatedAt:       createdAt,
	}, nil
}
