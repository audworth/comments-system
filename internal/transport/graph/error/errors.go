package error

type ErrorCode string

const (
	CodeInvalidArgument ErrorCode = "INVALID_ARGUMENT"
	CodeInvalidCursor   ErrorCode = "INVALID_CURSOR"
	CodeInvalidPageSize ErrorCode = "INVALID_PAGE_SIZE"

	CodeUnauthenticated ErrorCode = "UNAUTHENTICATED"
	CodeForbidden       ErrorCode = "FORBIDDEN"

	CodePostNotFound     ErrorCode = "POST_NOT_FOUND"
	CodeCommentNotFound  ErrorCode = "COMMENT_NOT_FOUND"
	CodeParentNotFound   ErrorCode = "PARENT_COMMENT_NOT_FOUND"
	CodeCommentsDisabled ErrorCode = "COMMENTS_DISABLED"

	CodePostTitleEmpty ErrorCode = "POST_TITLE_EMPTY"
	CodePostBodyEmpty  ErrorCode = "POST_BODY_EMPTY"

	CodeCommentEmpty   ErrorCode = "COMMENT_EMPTY"
	CodeCommentTooLong ErrorCode = "COMMENT_TOO_LONG"

	CodeRequestCancelled ErrorCode = "REQUEST_CANCELLED"
	CodeDeadlineExceeded ErrorCode = "DEADLINE_EXCEEDED"
	CodeInternal         ErrorCode = "INTERNAL"
)

type ClientError struct {
	Code    ErrorCode
	Message string
	Field   string
	Err     error
}

func (e *ClientError) Error() string {
	return e.Message
}

func (e *ClientError) Unwrap() error {
	return e.Err
}

func InvalidArgument(field string, message string, cause error) error {
	return &ClientError{
		Code:    CodeInvalidArgument,
		Message: message,
		Field:   field,
		Err:     cause,
	}
}

func InvalidID(field string, cause error) error {
	return InvalidArgument(field, "invalid ID", cause)
}

func InvalidCursor(field string, cause error) error {
	return &ClientError{
		Code:    CodeInvalidCursor,
		Message: "invalid cursor",
		Field:   field,
		Err:     cause,
	}
}
