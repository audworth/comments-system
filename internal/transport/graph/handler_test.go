package graph_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/audworth/comments-system/internal/application/comment"
	"github.com/audworth/comments-system/internal/application/post"
	"github.com/audworth/comments-system/internal/application/user"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/audworth/comments-system/internal/platform/db/inmem"
	"github.com/audworth/comments-system/internal/storage/mem"
	"github.com/audworth/comments-system/internal/transport/graph"
	"github.com/audworth/comments-system/internal/transport/graph/graphscalar"
	"github.com/audworth/comments-system/internal/transport/graph/resolver"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type noopEvents struct{}

func (noopEvents) NotifyCommentCreated(context.Context, *domain.Comment) error { return nil }

func (noopEvents) SubscribeCommentCreated(context.Context, uuid.UUID) (<-chan *domain.Comment, error) {
	comments := make(chan *domain.Comment)
	close(comments)
	return comments, nil
}

type testAPI struct {
	client *client.Client
	users  []domain.User
}

func newTestAPI(t *testing.T) testAPI {
	t.Helper()

	logger := slog.New(slog.DiscardHandler)
	db := inmem.New()
	users := []domain.User{
		{ID: uuid.New(), Name: "alice"},
		{ID: uuid.New(), Name: "bob"},
	}
	require.NoError(t, db.Update(func(tx *inmem.Tx) error {
		for _, u := range users {
			tx.PutUser(u)
		}
		return nil
	}))

	usersService := user.NewService(mem.NewUserRepository(db, logger), logger)
	postsService := post.NewService(mem.NewPostRepository(db, logger), logger)
	commentsService := comment.NewService(
		mem.NewCommentsRepository(db, logger),
		noopEvents{},
		noopEvents{},
		logger,
	)

	handler := graph.NewHandler(
		resolver.New(postsService, usersService, commentsService),
		usersService,
		commentsService,
		logger,
		graph.HandlerConfig{},
	)

	return testAPI{
		client: client.New(handler, client.Path("/query")),
		users:  users,
	}
}

type postResult struct {
	ID              string
	Title           string
	CommentsEnabled bool
}

func createPost(
	t *testing.T,
	api testAPI,
	authorID uuid.UUID,
	title string,
	commentsEnabled bool,
) postResult {
	t.Helper()

	var response struct {
		CreatePost postResult
	}
	err := api.client.Post(`
		mutation CreatePost($input: CreatePostInput!) {
			createPost(input: $input) {
				id
				title
				commentsEnabled
			}
		}
	`, &response, client.Var("input", map[string]any{
		"authorId":        authorID.String(),
		"title":           title,
		"body":            title + " body",
		"commentsEnabled": commentsEnabled,
	}))
	require.NoError(t, err)

	return response.CreatePost
}

func TestGraphQL_PostsCommentsAndRepliesWorkflow(t *testing.T) {
	t.Parallel()

	api := newTestAPI(t)
	post := createPost(t, api, api.users[0].ID, "first post", true)
	secondPost := createPost(t, api, api.users[1].ID, "second post", true)

	type createdComment struct {
		ID   string
		Body string
	}

	const createCommentMutation = `
		mutation CreateComment($input: CreateCommentInput!) {
			createComment(input: $input) { id body }
		}
	`
	rootBodies := []string{"root one", "root two", "root three"}
	roots := make([]createdComment, 0, len(rootBodies))
	for i, body := range rootBodies {
		var response struct {
			CreateComment createdComment
		}
		require.NoError(t, api.client.Post(createCommentMutation, &response, client.Var("input", map[string]any{
			"postId":   post.ID,
			"authorId": api.users[i%len(api.users)].ID.String(),
			"body":     body,
		})))
		roots = append(roots, response.CreateComment)
	}

	var replyResponse struct {
		CreateComment createdComment
	}
	require.NoError(t, api.client.Post(createCommentMutation, &replyResponse, client.Var("input", map[string]any{
		"postId":   post.ID,
		"parentId": roots[0].ID,
		"authorId": api.users[1].ID.String(),
		"body":     "nested reply",
	})))
	reply := replyResponse.CreateComment

	type pageInfo struct {
		EndCursor   *graphscalar.Cursor
		HasNextPage bool
	}
	type commentNode struct {
		ID     string
		Body   string
		Author struct {
			ID   string
			Name string
		}
	}

	var firstPage struct {
		Post struct {
			ID     string
			Title  string
			Author struct {
				ID   string
				Name string
			}
			Comments struct {
				Nodes    []commentNode
				PageInfo pageInfo
			}
		}
	}
	require.NoError(t, api.client.Post(`
		query PostWithComments($id: ID!) {
			post(id: $id) {
				id
				title
				author { id name }
				comments(first: 2) {
					nodes { id body author { id name } }
					pageInfo { endCursor hasNextPage }
				}
			}
		}
	`, &firstPage, client.Var("id", post.ID)))

	require.Equal(t, post.ID, firstPage.Post.ID)
	require.Equal(t, "first post", firstPage.Post.Title)
	require.Equal(t, api.users[0].ID.String(), firstPage.Post.Author.ID)
	require.Equal(t, "alice", firstPage.Post.Author.Name)
	require.Len(t, firstPage.Post.Comments.Nodes, 2)
	require.True(t, firstPage.Post.Comments.PageInfo.HasNextPage)
	require.NotNil(t, firstPage.Post.Comments.PageInfo.EndCursor)

	var nextPage struct {
		Comments struct {
			Nodes    []commentNode
			PageInfo pageInfo
		}
	}
	require.NoError(t, api.client.Post(`
		query MoreComments($postId: ID!, $after: Cursor!) {
			comments(postId: $postId, first: 2, after: $after) {
				nodes { id body author { id name } }
				pageInfo { endCursor hasNextPage }
			}
		}
	`, &nextPage,
		client.Var("postId", post.ID),
		client.Var("after", *firstPage.Post.Comments.PageInfo.EndCursor),
	))
	require.Len(t, nextPage.Comments.Nodes, 1)
	require.False(t, nextPage.Comments.PageInfo.HasNextPage)

	rootIDs := make([]string, 0, len(roots))
	for _, node := range firstPage.Post.Comments.Nodes {
		rootIDs = append(rootIDs, node.ID)
	}
	for _, node := range nextPage.Comments.Nodes {
		rootIDs = append(rootIDs, node.ID)
	}
	wantRootIDs := []string{roots[0].ID, roots[1].ID, roots[2].ID}
	require.ElementsMatch(t, wantRootIDs, rootIDs)
	require.NotContains(t, rootIDs, reply.ID)

	var thread struct {
		Comment struct {
			Replies struct {
				Nodes []commentNode
			}
		}
	}
	require.NoError(t, api.client.Post(`
		query Thread($id: ID!) {
			comment(id: $id) {
				replies(first: 10) { nodes { id body } }
			}
		}
	`, &thread, client.Var("id", roots[0].ID)))
	require.Equal(t, []commentNode{{ID: reply.ID, Body: "nested reply"}}, thread.Comment.Replies.Nodes)

	var postsPage struct {
		Posts struct {
			Nodes    []postResult
			PageInfo pageInfo
		}
	}
	require.NoError(t, api.client.Post(`
		query PostsPage { posts(first: 1) { nodes { id title } pageInfo { endCursor hasNextPage } } }
	`, &postsPage))
	require.Len(t, postsPage.Posts.Nodes, 1)
	require.True(t, postsPage.Posts.PageInfo.HasNextPage)
	require.NotNil(t, postsPage.Posts.PageInfo.EndCursor)

	var remainingPosts struct {
		Posts struct {
			Nodes []postResult
		}
	}
	require.NoError(t, api.client.Post(`
		query PostsPage($after: Cursor!) { posts(first: 1, after: $after) { nodes { id title } } }
	`, &remainingPosts, client.Var("after", *postsPage.Posts.PageInfo.EndCursor)))
	require.Len(t, remainingPosts.Posts.Nodes, 1)
	require.ElementsMatch(
		t,
		[]string{post.ID, secondPost.ID},
		[]string{postsPage.Posts.Nodes[0].ID, remainingPosts.Posts.Nodes[0].ID},
	)
}

func TestGraphQL_OnlyPostAuthorCanDisableComments(t *testing.T) {
	t.Parallel()

	api := newTestAPI(t)
	created := createPost(t, api, api.users[0].ID, "protected post", true)

	forbidden, err := api.client.RawPost(`
		mutation SetComments($input: SetPostCommentsEnabledInput!) {
			setPostCommentsEnabled(input: $input) { id }
		}
	`, client.Var("input", map[string]any{
		"postId":   created.ID,
		"authorId": api.users[1].ID.String(),
		"enabled":  false,
	}))
	require.NoError(t, err)
	requireGraphQLError(t, forbidden, "FORBIDDEN", "")

	var response struct {
		SetPostCommentsEnabled postResult
	}
	require.NoError(t, api.client.Post(`
		mutation SetComments($input: SetPostCommentsEnabledInput!) {
			setPostCommentsEnabled(input: $input) { id commentsEnabled }
		}
	`, &response, client.Var("input", map[string]any{
		"postId":   created.ID,
		"authorId": api.users[0].ID.String(),
		"enabled":  false,
	})))
	require.False(t, response.SetPostCommentsEnabled.CommentsEnabled)

	disabled, err := api.client.RawPost(`
		mutation CreateComment($input: CreateCommentInput!) {
			createComment(input: $input) { id }
		}
	`, client.Var("input", map[string]any{
		"postId":   created.ID,
		"authorId": api.users[0].ID.String(),
		"body":     "must be rejected",
	}))
	require.NoError(t, err)
	requireGraphQLError(t, disabled, "COMMENTS_DISABLED", "")
}

func TestGraphQL_ReportsValidationErrors(t *testing.T) {
	t.Parallel()

	api := newTestAPI(t)
	created := createPost(t, api, api.users[0].ID, "validation post", true)

	tests := []struct {
		name      string
		query     string
		options   []client.Option
		wantCode  string
		wantField string
	}{
		{
			name:      "некорректный идентификатор",
			query:     `query { post(id: "not-a-uuid") { id } }`,
			wantCode:  "INVALID_ARGUMENT",
			wantField: "postId",
		},
		{
			name:  "некорректный курсор",
			query: `query Comments($postId: ID!) { comments(postId: $postId, after: "not-base64") { nodes { id } } }`,
			options: []client.Option{
				client.Var("postId", created.ID),
			},
			wantCode:  "INVALID_CURSOR",
			wantField: "after",
		},
		{
			name:  "некорректный размер страницы",
			query: `query Comments($postId: ID!) { comments(postId: $postId, first: 0) { nodes { id } } }`,
			options: []client.Option{
				client.Var("postId", created.ID),
			},
			wantCode:  "INVALID_PAGE_SIZE",
			wantField: "first",
		},
		{
			name: "пустой комментарий",
			query: `
				mutation CreateComment($input: CreateCommentInput!) {
					createComment(input: $input) { id }
				}
			`,
			options: []client.Option{
				client.Var("input", map[string]any{
					"postId":   created.ID,
					"authorId": api.users[0].ID.String(),
					"body":     "   ",
				}),
			},
			wantCode:  "COMMENT_EMPTY",
			wantField: "body",
		},
		{
			name: "слишком длинный комментарий",
			query: `
				mutation CreateComment($input: CreateCommentInput!) {
					createComment(input: $input) { id }
				}
			`,
			options: []client.Option{
				client.Var("input", map[string]any{
					"postId":   created.ID,
					"authorId": api.users[0].ID.String(),
					"body":     strings.Repeat("я", domain.MaxCommentLength+1),
				}),
			},
			wantCode:  "COMMENT_TOO_LONG",
			wantField: "body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			response, err := api.client.RawPost(tt.query, tt.options...)
			require.NoError(t, err)
			requireGraphQLError(t, response, tt.wantCode, tt.wantField)
		})
	}
}

type graphQLError struct {
	Extensions struct {
		Code  string `json:"code"`
		Field string `json:"field"`
	} `json:"extensions"`
}

func requireGraphQLError(t *testing.T, response *client.Response, code string, field string) {
	t.Helper()

	var errors []graphQLError
	require.NoError(t, json.Unmarshal(response.Errors, &errors))
	require.Len(t, errors, 1)
	require.Equal(t, code, errors[0].Extensions.Code)
	require.Equal(t, field, errors[0].Extensions.Field)
}
