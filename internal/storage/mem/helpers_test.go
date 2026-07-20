package mem

import (
	"testing"
	"time"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/audworth/comments-system/internal/platform/db/inmem"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var testTime = time.Date(2026, time.July, 20, 12, 0, 0, 0, time.UTC)

func testID(value byte) uuid.UUID {
	var id uuid.UUID
	id[len(id)-1] = value
	return id
}

func newTestDB(
	t *testing.T,
	users []domain.User,
	posts []domain.Post,
	comments []domain.Comment,
) *inmem.InMem {
	t.Helper()

	db := inmem.New()
	err := db.Update(func(tx *inmem.Tx) error {
		for _, u := range users {
			tx.PutUser(u)
		}
		for _, p := range posts {
			tx.PutPost(p)
		}
		for _, comm := range comments {
			tx.PutComment(comm)
		}
		return nil
	})
	require.NoError(t, err)

	return db
}
