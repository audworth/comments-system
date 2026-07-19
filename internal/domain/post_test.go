package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNewPost(t *testing.T) {
	t.Parallel()

	id, authorID := uuid.New(), uuid.New()
	createdAt := time.Date(2026, time.July, 18, 12, 0, 0, 0, time.UTC)

	p, err := NewPost(id, authorID, "  заголовок\n", "\tтело\r\n", true, createdAt)

	require.NoError(t, err)
	require.Equal(t, id, p.ID)
	require.Equal(t, authorID, p.AuthorID)
	require.Equal(t, "заголовок", p.Title)
	require.Equal(t, "тело", p.Body)
	require.True(t, p.CommentsEnabled)
	require.Equal(t, createdAt, p.CreatedAt)
	require.Equal(t, createdAt, p.UpdatedAt)
}

func TestNewPost_RejectsEmptyContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		title   string
		body    string
		wantErr error
	}{
		{name: "пустой заголовок", title: "", body: "тело", wantErr: ErrEmptyPostTitle},
		{name: "whitespace заголовок", title: " \t\r\n", body: "тело", wantErr: ErrEmptyPostTitle},
		{name: "пустое тело", title: "заголовок", body: "", wantErr: ErrEmptyPostBody},
		{name: "whitespace тело", title: "заголовок", body: " \t\r\n", wantErr: ErrEmptyPostBody},
		{name: "сначала проверяется заголовок", title: "", body: "", wantErr: ErrEmptyPostTitle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewPost(uuid.New(), uuid.New(), tt.title, tt.body, false, time.Now())

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestNewPost_AcceptsLongUnicodeContent(t *testing.T) {
	t.Parallel()

	title := strings.Repeat("я", 10_000)
	body := strings.Repeat("💨", 10_000)

	post, err := NewPost(uuid.New(), uuid.New(), title, body, false, time.Now())

	require.NoError(t, err)
	require.Equal(t, title, post.Title)
	require.Equal(t, body, post.Body)
}
