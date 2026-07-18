package error

import (
	"context"
	"errors"
	"log/slog"
	"runtime/debug"

	"github.com/99designs/gqlgen/graphql"
)

var ErrRecovered = errors.New("recovered graphql panic")

func NewRecoverFunc(logger *slog.Logger) func(context.Context, any) error {
	return func(ctx context.Context, panicInfo any) error {
		logger.ErrorContext(
			ctx,
			"panic during graphql execution",
			slog.Any("panic", panicInfo),
			slog.Any("path", graphql.GetPath(ctx)),
			slog.String("stack", string(debug.Stack())),
		)

		return ErrRecovered
	}
}
