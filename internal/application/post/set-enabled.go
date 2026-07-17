package post

import (
	"context"
	"fmt"

	"github.com/audworth/comments-system/internal/domain"
	"github.com/google/uuid"
)

func (s *Service) SetCommentsToEnabled(ctx context.Context, postID uuid.UUID, author uuid.UUID, enabled bool) (*domain.Post, error) {
	post, err := s.repo.SetCommentsEnabled(ctx, postID, author, enabled)
	if err != nil {
		return nil, fmt.Errorf("set comments enabled for post %s (author %s): %w", postID, author, err)
	}

	return post, nil
}
