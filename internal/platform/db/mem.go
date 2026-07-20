package db

import (
	"sync"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

type InMemory struct {
	mu sync.RWMutex

	users    map[uuid.UUID]domain.User
	posts    map[uuid.UUID]domain.Post
	comments map[uuid.UUID]domain.Comment

	postIDs            []uuid.UUID
	commentIDsByParent map[commentListKey][]uuid.UUID
}

func NewInMemory() *InMemory {
	return &InMemory{
		users:              make(map[uuid.UUID]domain.User),
		posts:              make(map[uuid.UUID]domain.Post),
		comments:           make(map[uuid.UUID]domain.Comment),
		postIDs:            make([]uuid.UUID, 0),
		commentIDsByParent: make(map[commentListKey][]uuid.UUID),
	}
}

type commentListKey struct {
	PostID   uuid.UUID
	ParentID uuid.UUID
}
