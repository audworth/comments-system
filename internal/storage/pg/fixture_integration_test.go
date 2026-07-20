package pg

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	dbURL    string
	testTime = time.Date(2026, time.July, 20, 12, 0, 0, 0, time.UTC)
)

func TestMain(m *testing.M) {
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		os.Exit(m.Run())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	cont, err := tcpostgres.Run(
		ctx,
		"postgres:18.4-alpine3.24",
		tcpostgres.WithDatabase("comments_test"),
		tcpostgres.WithUsername("comments"),
		tcpostgres.WithPassword("comments"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		cancel()
		log.Fatalf("запуск PostgreSQL testcontainer: %v\n", err)
	}

	dbURL, err = cont.ConnectionString(ctx, "sslmode=disable")
	cancel()
	if err != nil {
		_ = testcontainers.TerminateContainer(cont)
		log.Fatalf("получение PostgreSQL connection string: %v\n", err)
	}

	code := m.Run()
	if err := testcontainers.TerminateContainer(cont); err != nil {
		fmt.Printf("остановка PostgreSQL testcontainer: %v\n", err)
		if code == 0 {
			code = 1
		}
	}
	os.Exit(code)
}

type pgFixture struct {
	posts    *PostRepository
	comments *CommentsRepository
	user     domain.User
}

func newPGFixture(t *testing.T) *pgFixture {
	t.Helper()
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("INTEGRATION_TESTS skipped")
	}

	adminConfig, err := pgxpool.ParseConfig(dbURL)
	require.NoError(t, err)
	admin, err := pgxpool.NewWithConfig(t.Context(), adminConfig)
	require.NoError(t, err)
	require.NoError(t, admin.Ping(t.Context()))

	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	identifier := pgx.Identifier{schema}.Sanitize()
	_, err = admin.Exec(t.Context(), "create schema "+identifier)
	require.NoError(t, err)

	poolConfig, err := pgxpool.ParseConfig(dbURL)
	require.NoError(t, err)
	poolConfig.ConnConfig.RuntimeParams["search_path"] = schema
	poolConfig.AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
		conn.TypeMap().RegisterType(&pgtype.Type{
			Name:  "timestamptz",
			OID:   pgtype.TimestamptzOID,
			Codec: &pgtype.TimestamptzCodec{ScanLocation: time.UTC},
		})
		return nil
	}
	pool, err := pgxpool.NewWithConfig(t.Context(), poolConfig)
	require.NoError(t, err)
	require.NoError(t, pool.Ping(t.Context()))

	t.Cleanup(func() {
		pool.Close()
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = admin.Exec(cleanupCtx, "drop schema "+identifier+" cascade")
		admin.Close()
	})

	applyTestMigrations(t, pool)

	user := domain.User{ID: uuid.New(), Name: "test_user"}
	_, err = pool.Exec(t.Context(), `insert into users (id, name) values ($1, $2)`, user.ID, user.Name)
	require.NoError(t, err)

	logger := slog.New(slog.DiscardHandler)
	return &pgFixture{
		posts:    NewPostRepository(pool, logger),
		comments: NewCommentsRepository(pool, logger),
		user:     user,
	}
}

func applyTestMigrations(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	migrationsDir := filepath.Join(filepath.Dir(filename), "..", "..", "..", "migrations")

	for _, name := range []string{
		"000001_create_users_table.up.sql",
		"000002_create_posts_table.up.sql",
		"000003_create_comments_table.up.sql",
		"000004_add_posts_created_at_index.up.sql",
		"000005_add_comments_post_parent_created_id_index.up.sql",
	} {
		migration, err := os.ReadFile(filepath.Join(migrationsDir, name))
		require.NoError(t, err)
		_, err = pool.Exec(t.Context(), string(migration))
		require.NoError(t, err)
	}
}

func publishTestPost(
	t *testing.T,
	db *pgFixture,
	id uuid.UUID,
	commentsEnabled bool,
	createdAt time.Time,
) domain.Post {
	t.Helper()

	want := domain.Post{
		ID:              id,
		AuthorID:        db.user.ID,
		Title:           "post " + id.String(),
		Body:            "body",
		CommentsEnabled: commentsEnabled,
		CreatedAt:       createdAt,
		UpdatedAt:       createdAt,
	}
	created, err := db.posts.Publish(t.Context(), &want)
	require.NoError(t, err)
	require.Equal(t, want, *created)
	return want
}

func testID(last byte) uuid.UUID {
	var id uuid.UUID
	id[len(id)-1] = last
	return id
}
