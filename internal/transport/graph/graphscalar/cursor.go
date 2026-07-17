package graphscalar

import (
	"errors"
	"io"
	"strconv"
)

type Cursor string

var ErrInvalidCursorType = errors.New("cursor must be a string")

func (c Cursor) MarshalGQL(w io.Writer) {
	_, _ = io.WriteString(w, strconv.Quote(string(c)))
}

func (c *Cursor) UnmarshalGQL(value any) error {
	raw, ok := value.(string)
	if !ok {
		return ErrInvalidCursorType
	}

	*c = Cursor(raw)
	return nil
}
