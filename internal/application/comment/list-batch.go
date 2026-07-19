package comment

import (
	"context"
	"fmt"

	"github.com/audworth/comments-system/internal/application"
)

func (s *Service) ListBatchComments(
	ctx context.Context,
	params []ListParams,
) ([]*Page, error) {
	if len(params) == 0 {
		return []*Page{}, nil
	}

	for i := range params {
		if params[i].Limit < 1 || params[i].Limit > 100 {
			return nil, fmt.Errorf("invalid comment page size %d: %w", i, application.ErrInvalidPageSize)
		}
	}

	pages, err := s.repo.ListChildrenBatch(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list comment pages: %w", err)
	}

	if len(pages) != len(params) {
		return nil, fmt.Errorf(
			"list comment pages: returned %d pages for %d requests",
			len(pages),
			len(params),
		)
	}

	return pages, nil
}
