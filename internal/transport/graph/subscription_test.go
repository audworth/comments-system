package graph_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/99designs/gqlgen/client"
	"github.com/audworth/comments-system/internal/application/comment"
	"github.com/audworth/comments-system/internal/application/post"
	"github.com/audworth/comments-system/internal/application/user"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/audworth/comments-system/internal/notifier"
	"github.com/audworth/comments-system/internal/platform/db/inmem"
	"github.com/audworth/comments-system/internal/platform/redis"
	"github.com/audworth/comments-system/internal/storage/mem"
	"github.com/audworth/comments-system/internal/transport/graph"
	"github.com/audworth/comments-system/internal/transport/graph/resolver"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestGraphQLSubscriptionIntegration_DeliversOnlyPostComments(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("INTEGRATION_TESTS не равен true")
	}

	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()

	cont, err := testcontainers.Run(
		ctx,
		"redis:8.8.0-alpine3.23",
		testcontainers.WithExposedPorts("6379/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").WithStartupTimeout(time.Minute),
		),
	)
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, cont)

	url, err := cont.Endpoint(ctx, "redis")
	require.NoError(t, err)
	redisClient, err := redis.NewClient(ctx, url)
	require.NoError(t, err)
	t.Cleanup(func() { _ = redisClient.Close() })

	logger := slog.New(slog.DiscardHandler)
	db := inmem.New()
	author := domain.User{ID: uuid.New(), Name: "author"}
	require.NoError(t, db.Update(func(tx *inmem.Tx) error {
		tx.PutUser(author)
		return nil
	}))

	usersService := user.NewService(mem.NewUserRepository(db, logger), logger)
	postsService := post.NewService(mem.NewPostRepository(db, logger), logger)
	events := notifier.NewNotifier(redisClient, logger)
	subscriber, err := notifier.NewSubscriber(ctx, redisClient, logger)
	require.NoError(t, err)
	t.Cleanup(subscriber.Close)
	commentsService := comment.NewService(
		mem.NewCommentsRepository(db, logger),
		events,
		subscriber,
		logger,
	)

	subscribedPost, err := postsService.Publish(ctx, post.PublishParams{
		AuthorID:        author.ID,
		Title:           "subscribed post",
		Body:            "body",
		CommentsEnabled: true,
	})
	require.NoError(t, err)
	otherPost, err := postsService.Publish(ctx, post.PublishParams{
		AuthorID:        author.ID,
		Title:           "other post",
		Body:            "body",
		CommentsEnabled: true,
	})
	require.NoError(t, err)

	redisSubscribers, err := redisClient.PubSubNumSub(ctx, "comment.created").Result()
	require.NoError(t, err)
	require.EqualValues(t, 1, redisSubscribers["comment.created"])

	handler := graph.NewHandler(
		resolver.New(postsService, usersService, commentsService),
		usersService,
		commentsService,
		logger,
		graph.HandlerConfig{},
	)
	graphClient := client.New(handler, client.Path("/query"))
	subscription := graphClient.Websocket(`
		subscription CommentCreated($postId: ID!) {
			commentCreated(postId: $postId) {
				id
				postId
				authorId
				body
			}
		}
	`, client.Var("postId", subscribedPost.ID.String()))
	t.Cleanup(func() { _ = subscription.Close() })

	type createdComment struct {
		ID       string
		PostID   string
		AuthorID string
		Body     string
	}
	createComment := func(postID uuid.UUID, body string) createdComment {
		t.Helper()

		var response struct {
			CreateComment createdComment
		}
		err := graphClient.Post(`
			mutation CreateComment($input: CreateCommentInput!) {
				createComment(input: $input) { id postId authorId body }
			}
		`, &response, client.Var("input", map[string]any{
			"postId":   postID.String(),
			"authorId": author.ID.String(),
			"body":     body,
		}))
		require.NoError(t, err)
		return response.CreateComment
	}

	_ = createComment(otherPost.ID, "other comment")
	want := createComment(subscribedPost.ID, "subscribed comment")

	var received struct {
		CommentCreated createdComment
	}
	next := make(chan error, 1)
	go func() {
		next <- subscription.Next(&received)
	}()

	select {
	case err := <-next:
		require.NoError(t, err)
		require.Equal(t, want, received.CommentCreated)
	case <-time.After(20 * time.Second):
		t.Fatal("событие о создании комментария не получено")
	}

	subCtx, stopSub := context.WithCancel(t.Context())
	comments, err := subscriber.SubscribeCommentCreated(subCtx, subscribedPost.ID)
	require.NoError(t, err)
	stopSub()

	select {
	case _, open := <-comments:
		require.False(t, open)
	case <-time.After(3 * time.Second):
		t.Fatal("локальная подписка не закрылась после отмены контекста")
	}
}
