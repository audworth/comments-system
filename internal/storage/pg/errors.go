package pg

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	foreignKeyViolationCode  = "23503"
	postsAuthorConstraint    = "posts_author_id_fkey"
	commentsAuthorConstraint = "comments_author_id_fkey"
)

func isForeignKeyViolation(err error, constraint string) bool {
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	return ok &&
		pgErr.Code == foreignKeyViolationCode &&
		pgErr.ConstraintName == constraint
}
