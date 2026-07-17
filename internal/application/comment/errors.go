package comment

import "errors"

var (
	ErrNotFound         = errors.New("comment not found")
	ErrPostNotFound     = errors.New("post not found")
	ErrParentNotFound   = errors.New("parent comment not found")
	ErrCommentsDisabled = errors.New("comments are disabled")
	ErrInvalidPageSize  = errors.New("invalid page size")
)
