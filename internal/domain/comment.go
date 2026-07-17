package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type Comment struct {
	ID        uuid.UUID
	PostID    uuid.UUID
	ParentID  *uuid.UUID
	AuthorID  uuid.UUID
	Body      string
	CreatedAt time.Time
}

func NewComment(
	id uuid.UUID,
	postID uuid.UUID,
	parentID *uuid.UUID,
	authorID uuid.UUID,
	body string,
	now time.Time,
) (Comment, error) {
	body = strings.TrimSpace(body)

	if body == "" {
		return Comment{}, ErrEmptyComment
	}

	if utf8.RuneCountInString(body) > 2000 {
		return Comment{}, ErrCommentTooLong
	}

	if parentID != nil && *parentID == id {
		return Comment{}, ErrSelfParent
	}

	return Comment{
		ID:        id,
		PostID:    postID,
		ParentID:  parentID,
		AuthorID:  authorID,
		Body:      body,
		CreatedAt: now,
	}, nil
}
