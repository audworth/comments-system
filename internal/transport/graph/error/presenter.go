package error

import (
	"context"
	"errors"
	"log/slog"

	"github.com/99designs/gqlgen/graphql"
	"github.com/audworth/comments-system/internal/application"
	"github.com/audworth/comments-system/internal/application/comment"
	"github.com/audworth/comments-system/internal/application/post"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type Presenter struct {
	logger *slog.Logger
}

func NewPresenter(logger *slog.Logger) *Presenter {
	return &Presenter{
		logger: logger,
	}
}

func (p *Presenter) Present(ctx context.Context, err error) *gqlerror.Error {
	presentedErr := graphql.DefaultErrorPresenter(ctx, err)

	if clientErr, ok := errors.AsType[*ClientError](err); ok {
		return apiError(
			presentedErr,
			clientErr.Code,
			clientErr.Message,
			clientErr.Field,
		)
	}

	switch {
	case errors.Is(err, domain.ErrEmptyPostTitle):
		return apiError(
			presentedErr,
			CodePostTitleEmpty,
			"post title must not be empty",
			"title",
		)
	case errors.Is(err, domain.ErrEmptyPostBody):
		return apiError(
			presentedErr,
			CodePostBodyEmpty,
			"post body must not be empty",
			"body",
		)
	case errors.Is(err, domain.ErrEmptyComment):
		return apiError(
			presentedErr,
			CodeCommentEmpty,
			"comment must not be empty",
			"body",
		)
	case errors.Is(err, domain.ErrCommentTooLong):
		return apiError(
			presentedErr,
			CodeCommentTooLong,
			"comment must not exceed 2000 characters",
			"body",
		)
	case errors.Is(err, post.ErrNotFound),
		errors.Is(err, comment.ErrPostNotFound):
		return apiError(
			presentedErr,
			CodePostNotFound,
			"post not found",
			"",
		)

	case errors.Is(err, post.ErrForbidden):
		return apiError(
			presentedErr,
			CodeForbidden,
			"not allowed to perform this operation",
			"",
		)

	case errors.Is(err, comment.ErrNotFound):
		return apiError(
			presentedErr,
			CodeCommentNotFound,
			"comment not found",
			"",
		)

	case errors.Is(err, comment.ErrParentNotFound):
		return apiError(
			presentedErr,
			CodeParentNotFound,
			"parent comment not found",
			"parentId",
		)

	case errors.Is(err, comment.ErrCommentsDisabled):
		return apiError(
			presentedErr,
			CodeCommentsDisabled,
			"comments are disabled for this post",
			"",
		)

	case errors.Is(err, application.ErrInvalidPageSize):
		return apiError(
			presentedErr,
			CodeInvalidPageSize,
			"page size must be between 1 and 100",
			"first",
		)

	case errors.Is(err, context.Canceled):
		return apiError(
			presentedErr,
			CodeRequestCancelled,
			"request was cancelled",
			"",
		)

	case errors.Is(err, context.DeadlineExceeded):
		return apiError(
			presentedErr,
			CodeDeadlineExceeded,
			"request deadline exceeded",
			"",
		)
	}

	if _, ok := errors.AsType[*gqlerror.Error](err); ok {
		return presentedErr
	}

	p.logger.ErrorContext(ctx, "graphql error", slog.Any("error", err), slog.Any("path", graphql.GetPath(ctx)))

	return apiError(
		presentedErr,
		CodeInternal,
		"internal server error",
		"",
	)
}

func apiError(err *gqlerror.Error, code ErrorCode, message string, field string) *gqlerror.Error {
	err.Message = message
	err.Extensions = map[string]any{
		"code": string(code),
	}

	if field != "" {
		err.Extensions["field"] = field
	}

	return err
}
