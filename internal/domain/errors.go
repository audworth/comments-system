package domain

import "errors"

var (
	ErrEmptyComment     = errors.New("comment is empty")
	ErrCommentTooLong   = errors.New("comment is too long")
	ErrSelfParent       = errors.New("comment cannot reference itself")
	ErrNotPostAuthor    = errors.New("user is not the post author")
	ErrCommentsDisabled = errors.New("comments are disabled")

	ErrEmptyPostTitle = errors.New("post title must not be empty")
	ErrEmptyPostBody  = errors.New("post body must not be empty")
)
