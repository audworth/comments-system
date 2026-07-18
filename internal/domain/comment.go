package domain

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

var (
	ErrEmptyComment   = errors.New("comment is empty")
	ErrCommentTooLong = errors.New("comment is too long")
	ErrSelfParent     = errors.New("comment cannot reference itself")
)

const MaxCommentLength = 2000

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
	createdAt time.Time,
) (*Comment, error) {
	body = strings.TrimSpace(body)

	if body == "" {
		return &Comment{}, ErrEmptyComment
	}

	if utf8.RuneCountInString(body) > MaxCommentLength {
		return &Comment{}, ErrCommentTooLong
	}

	if parentID != nil && *parentID == id {
		return &Comment{}, ErrSelfParent
	}

	return &Comment{
		ID:        id,
		PostID:    postID,
		ParentID:  parentID,
		AuthorID:  authorID,
		Body:      body,
		CreatedAt: createdAt,
	}, nil
}
