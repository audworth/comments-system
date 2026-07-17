package domain

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID              uuid.UUID
	AuthorID        uuid.UUID
	Title           string
	Body            string
	CommentsEnabled bool
	CreatedAt       time.Time
}

func (p *Post) SetCommentEnabled(userID uuid.UUID, enabled bool) error {
	if p.AuthorID != userID {
		return ErrNotPostAuthor
	}

	p.CommentsEnabled = enabled

	return nil
}
