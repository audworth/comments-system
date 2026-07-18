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

	id := uuid.New()
	author := User{ID: uuid.New(), Name: "автор"}
	createdAt := time.Date(2026, time.July, 18, 12, 0, 0, 0, time.UTC)

	p, err := NewPost(id, author, "  заголовок\n", "\tтело\r\n", true, createdAt)

	require.NoError(t, err)
	require.Equal(t, id, p.ID)
	require.Equal(t, author, p.Author)
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

			_, err := NewPost(uuid.New(), User{ID: uuid.New()}, tt.title, tt.body, false, time.Now())

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestNewPost_AcceptsLongUnicodeContent(t *testing.T) {
	t.Parallel()

	title := strings.Repeat("я", 10_000)
	body := strings.Repeat("💨", 10_000)

	post, err := NewPost(uuid.New(), User{ID: uuid.New()}, title, body, false, time.Now())

	require.NoError(t, err)
	require.Equal(t, title, post.Title)
	require.Equal(t, body, post.Body)
}

func TestPost_SetCommentEnabled(t *testing.T) {
	t.Parallel()

	names := map[bool]string{false: "disabled", true: "enabled"}
	for _, enabled := range []bool{false, true} {
		t.Run(names[enabled], func(t *testing.T) {
			t.Parallel()

			authorID := uuid.New()
			p := Post{
				Author:          User{ID: authorID},
				CommentsEnabled: !enabled,
			}

			err := p.SetCommentEnabled(authorID, enabled)

			require.NoError(t, err)
			require.Equal(t, enabled, p.CommentsEnabled)
		})
	}
}

func TestPost_SetCommentEnabled_RejectsNonAuthor(t *testing.T) {
	t.Parallel()

	p := Post{
		Author:          User{ID: uuid.New()},
		CommentsEnabled: true,
	}

	err := p.SetCommentEnabled(uuid.New(), false)

	require.ErrorIs(t, err, ErrNotPostAuthor)
	require.True(t, p.CommentsEnabled)
}
