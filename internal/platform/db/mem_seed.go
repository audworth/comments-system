package db

import (
	"fmt"
	"time"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

const (
	usersAmount       = 10
	postsAmount       = 10
	rootComments      = 1000
	repliesPerComment = 100
	maxThreadDepth    = 100
	totalComments     = postsAmount*rootComments*(repliesPerComment+1) + maxThreadDepth + 1
)

func (db *InMemory) Seed() {
	seeded := &InMemory{
		users:              make(map[uuid.UUID]domain.User),
		posts:              make(map[uuid.UUID]domain.Post),
		comments:           make(map[uuid.UUID]domain.Comment, totalComments),
		postIDs:            make([]uuid.UUID, 0),
		commentIDsByParent: make(map[commentListKey][]uuid.UUID),
	}

	users := make([]uuid.UUID, usersAmount)
	posts := make([]uuid.UUID, postsAmount)
	roots := make([][]uuid.UUID, postsAmount)

	for i := 1; i <= usersAmount; i++ {
		id := uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0000-%012d", i))
		users[i-1] = id
		seeded.users[id] = domain.User{
			ID:   id,
			Name: fmt.Sprintf("user_%d", i),
		}
	}

	for i := 1; i <= postsAmount; i++ {
		id := uuid.MustParse(fmt.Sprintf("10000000-0000-0000-0000-%012d", i))
		createdAt := time.Date(2026, time.February, 1+i, 0, 0, 0, 0, time.UTC)
		posts[i-1] = id
		seeded.posts[id] = domain.Post{
			ID:              id,
			AuthorID:        users[i-1],
			Title:           fmt.Sprintf("пост_%d", i),
			Body:            fmt.Sprintf("тело_поста_%d", i),
			CommentsEnabled: true,
			CreatedAt:       createdAt,
			UpdatedAt:       createdAt,
		}
		seeded.postIDs = append(seeded.postIDs, id)
		roots[i-1] = make([]uuid.UUID, rootComments)
	}

	for i := 1; i <= postsAmount; i++ {
		postID := posts[i-1]
		root := commentListKey{PostID: postID}
		seeded.commentIDsByParent[root] = make([]uuid.UUID, 0, rootComments+1)

		for j := 1; j <= rootComments; j++ {
			id := uuid.MustParse(fmt.Sprintf(
				"20000000-%04d-0000-0000-%012d",
				i,
				j,
			))

			createdAt := time.Date(2026, time.March, 1+i, 0, 0, j, 0, time.UTC)
			comm := domain.Comment{
				ID:        id,
				PostID:    postID,
				AuthorID:  users[(i+j-2)%usersAmount],
				Body:      fmt.Sprintf("comment_%d_%d", i, j),
				CreatedAt: createdAt,
			}

			roots[i-1][j-1] = id
			seeded.comments[id] = comm
			seeded.commentIDsByParent[root] = append(seeded.commentIDsByParent[root], id)
		}
	}

	for p := 1; p <= postsAmount; p++ {
		postID := posts[p-1]

		for c := 1; c <= rootComments; c++ {
			parentID := roots[p-1][c-1]
			key := commentListKey{PostID: postID, ParentID: parentID}
			seeded.commentIDsByParent[key] = make([]uuid.UUID, 0, repliesPerComment)

			for r := 1; r <= repliesPerComment; r++ {
				id := uuid.MustParse(fmt.Sprintf(
					"30000000-%04d-0000-0000-%012d",
					p,
					c*repliesPerComment+r,
				))
				createdAt := time.Date(2026, time.April, 1+p, 0, 0, c, r*1000, time.UTC)
				comm := domain.Comment{
					ID:        id,
					PostID:    postID,
					ParentID:  &parentID,
					AuthorID:  users[(p+c+r-3)%usersAmount],
					Body:      fmt.Sprintf("reply_%d_%d_%d", p, c, r),
					CreatedAt: createdAt,
				}

				seeded.comments[id] = comm
				seeded.commentIDsByParent[key] = append(seeded.commentIDsByParent[key], id)
			}
		}
	}

	postID := posts[0]
	rootID := uuid.MustParse("40000000-0001-0000-0000-000000000000")
	rootCreatedAt := time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC)
	seeded.comments[rootID] = domain.Comment{
		ID:        rootID,
		PostID:    postID,
		AuthorID:  users[0],
		Body:      "deep_thread_root",
		CreatedAt: rootCreatedAt,
	}
	rootKey := commentListKey{PostID: postID}
	seeded.commentIDsByParent[rootKey] = append(seeded.commentIDsByParent[rootKey], rootID)

	parentID := rootID
	for d := 1; d <= maxThreadDepth; d++ {
		id := uuid.MustParse(fmt.Sprintf("40000000-0001-0000-0000-%012d", d))
		commParentID := parentID
		comm := domain.Comment{
			ID:        id,
			PostID:    postID,
			ParentID:  &commParentID,
			AuthorID:  users[(d-1)%usersAmount],
			Body:      fmt.Sprintf("deep_reply_%d", d),
			CreatedAt: rootCreatedAt.Add(time.Duration(d) * time.Microsecond),
		}
		key := commentListKey{PostID: postID, ParentID: parentID}
		seeded.comments[id] = comm
		seeded.commentIDsByParent[key] = append(seeded.commentIDsByParent[key], id)
		parentID = id
	}

	db.mu.Lock()
	db.users = seeded.users
	db.posts = seeded.posts
	db.comments = seeded.comments
	db.postIDs = seeded.postIDs
	db.commentIDsByParent = seeded.commentIDsByParent
	db.mu.Unlock()
}
