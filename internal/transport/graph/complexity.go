package graph

import (
	"github.com/audworth/comments-system/internal/transport/graph/generated"
	"github.com/audworth/comments-system/internal/transport/graph/graphscalar"
)

const maxPageSize int32 = 100

func calculatePageComplexity(childComplexity int, first int32) int {
	if first < 1 {
		first = 1
	}
	if first > maxPageSize {
		first = maxPageSize
	}

	return 1 + int(first)*childComplexity
}

func configureQueryComplexity(config *generated.Config) {
	config.Complexity.Query.Posts = func(
		childComplexity int,
		first int32,
		_ *graphscalar.Cursor,
	) int {
		return calculatePageComplexity(childComplexity, first)
	}

	config.Complexity.Query.Comments = func(
		childComplexity int,
		_ string,
		_ *string,
		first int32,
		_ *graphscalar.Cursor,
	) int {
		return calculatePageComplexity(childComplexity, first)
	}

	config.Complexity.Post.Comments = func(
		childComplexity int,
		first int32,
		_ *graphscalar.Cursor,
	) int {
		return calculatePageComplexity(childComplexity, first)
	}

	config.Complexity.Comment.Replies = func(
		childComplexity int,
		first int32,
		_ *graphscalar.Cursor,
	) int {
		return calculatePageComplexity(childComplexity, first)
	}
}
