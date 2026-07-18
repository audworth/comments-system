package post

import "errors"

var (
	ErrNotFound  = errors.New("post not found")
	ErrForbidden = errors.New("operation is forbidden")
)
