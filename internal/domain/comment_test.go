package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNewComment(t *testing.T) {
	t.Parallel()

	id, postID, parentID, authorID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	createdAt := time.Date(2026, time.July, 18, 12, 0, 0, 0, time.UTC)

	comment, err := NewComment(id, postID, &parentID, authorID, "  комментарий\r\n", createdAt)

	require.NoError(t, err)
	require.Equal(t, id, comment.ID)
	require.Equal(t, postID, comment.PostID)
	require.NotNil(t, comment.ParentID)
	require.Equal(t, parentID, *comment.ParentID)
	require.Equal(t, authorID, comment.AuthorID)
	require.Equal(t, "комментарий", comment.Body)
	require.Equal(t, createdAt, comment.CreatedAt)
}

func TestNewComment_AcceptsMaximumBodyLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{name: "ascii", body: strings.Repeat("a", 2000)},
		{name: "utf-8", body: strings.Repeat("я", 2000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			comment, err := NewComment(uuid.New(), uuid.New(), nil, uuid.New(), tt.body, time.Now())

			require.NoError(t, err)
			require.Equal(t, tt.body, comment.Body)
		})
	}
}

func TestNewComment_RejectsInvalidBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		body    string
		wantErr error
	}{
		{name: "пустой", body: "", wantErr: ErrEmptyComment},
		{name: "whitespaces", body: " \t\r\n", wantErr: ErrEmptyComment},
		{name: "2001 ascii", body: strings.Repeat("a", 2001), wantErr: ErrCommentTooLong},
		{name: "2001 utf-8", body: strings.Repeat("я", 2001), wantErr: ErrCommentTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewComment(uuid.New(), uuid.New(), nil, uuid.New(), tt.body, time.Now())

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestNewComment_RejectsSelfParent(t *testing.T) {
	t.Parallel()

	id := uuid.New()

	_, err := NewComment(id, uuid.New(), &id, uuid.New(), "комментарий", time.Now())

	require.ErrorIs(t, err, ErrSelfParent)
}
