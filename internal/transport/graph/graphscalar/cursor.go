package graphscalar

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/google/uuid"
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

type payload struct {
	Version   int    `json:"v"`
	CreatedAt string `json:"createdAt"`
	ID        string `json:"id"`
}

type Position struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

func EncodeCursor(createdAt time.Time, id uuid.UUID) (Cursor, error) {
	data, err := json.Marshal(payload{
		CreatedAt: createdAt.UTC().Format(time.RFC3339Nano),
		ID:        id.String(),
	})
	if err != nil {
		return "", fmt.Errorf("marshal cursor: %w", err)
	}

	return Cursor(
		base64.RawURLEncoding.EncodeToString(data),
	), nil
}

func DecodeCursor(cursor Cursor) (Position, error) {
	raw, err := base64.RawURLEncoding.DecodeString(string(cursor))
	if err != nil {
		return Position{}, fmt.Errorf("decode base64: %w", err)
	}

	var value payload
	if err := json.Unmarshal(raw, &value); err != nil {
		return Position{}, fmt.Errorf("decode payload: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339Nano, value.CreatedAt)
	if err != nil {
		return Position{}, fmt.Errorf("parse cursor time: %w", err)
	}

	id, err := uuid.Parse(value.ID)
	if err != nil {
		return Position{}, fmt.Errorf("parse cursor id: %w", err)
	}

	return Position{
		CreatedAt: createdAt,
		ID:        id,
	}, nil
}
