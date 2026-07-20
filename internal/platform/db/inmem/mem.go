package inmem

import (
	"bytes"
	"slices"
	"sync"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

type InMem struct {
	mu sync.RWMutex

	users    map[uuid.UUID]domain.User
	posts    map[uuid.UUID]domain.Post
	comments map[uuid.UUID]domain.Comment

	postIDs            []uuid.UUID
	commentIDsByParent map[commentListKey][]uuid.UUID
}

func New() *InMem {
	return &InMem{
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

// https://dgraph-io.github.io/badger/quickstart.html
type Tx struct {
	db *InMem
}

func (db *InMem) View(fn func(*Tx) error) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return fn(&Tx{db: db})
}

func (db *InMem) Update(fn func(*Tx) error) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	return fn(&Tx{db: db})
}

func (tx *Tx) User(id uuid.UUID) (domain.User, bool) {
	user, ok := tx.db.users[id]
	return user, ok
}

func (tx *Tx) PutUser(user domain.User) {
	tx.db.users[user.ID] = user
}

func (tx *Tx) Post(id uuid.UUID) (domain.Post, bool) {
	post, ok := tx.db.posts[id]
	return post, ok
}

func (tx *Tx) PostIDs() []uuid.UUID {
	return tx.db.postIDs
}

func (tx *Tx) Comment(id uuid.UUID) (domain.Comment, bool) {
	comment, ok := tx.db.comments[id]
	return comment, ok
}

func (tx *Tx) CommentIDs(postID uuid.UUID, parentID *uuid.UUID) []uuid.UUID {
	key := commentListKey{PostID: postID}
	if parentID != nil {
		key.ParentID = *parentID
	}

	return tx.db.commentIDsByParent[key]
}

func (tx *Tx) PutPost(post domain.Post) {
	if _, exists := tx.db.posts[post.ID]; exists {
		tx.db.posts[post.ID] = post
		return
	}

	tx.db.posts[post.ID] = post
	index, _ := slices.BinarySearchFunc(tx.db.postIDs, post, func(id uuid.UUID, target domain.Post) int {
		return cmpPosts(tx.db.posts[id], target)
	})
	tx.db.postIDs = slices.Insert(tx.db.postIDs, index, post.ID)
}

func (tx *Tx) PutComment(comment domain.Comment) {
	tx.db.comments[comment.ID] = comment
	key := commentListKey{PostID: comment.PostID}
	if comment.ParentID != nil {
		key.ParentID = *comment.ParentID
	}

	ids := tx.db.commentIDsByParent[key]
	index, _ := slices.BinarySearchFunc(ids, comment, func(id uuid.UUID, target domain.Comment) int {
		return cmpComments(tx.db.comments[id], target)
	})
	tx.db.commentIDsByParent[key] = slices.Insert(ids, index, comment.ID)
}

func (db *InMem) sortIndexes() {
	slices.SortFunc(db.postIDs, func(a, b uuid.UUID) int {
		return cmpPosts(db.posts[a], db.posts[b])
	})

	for _, ids := range db.commentIDsByParent {
		slices.SortFunc(ids, func(a, b uuid.UUID) int {
			return cmpComments(db.comments[a], db.comments[b])
		})
	}
}

func cmpPosts(a, b domain.Post) int {
	if order := b.CreatedAt.Compare(a.CreatedAt); order != 0 {
		return order
	}

	return bytes.Compare(b.ID[:], a.ID[:])
}

func cmpComments(a, b domain.Comment) int {
	if order := b.CreatedAt.Compare(a.CreatedAt); order != 0 {
		return order
	}

	return bytes.Compare(b.ID[:], a.ID[:])
}
