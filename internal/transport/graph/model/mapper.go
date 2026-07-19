package model

import (
	"fmt"

	"github.com/audworth/comments-system/internal/application/comment"
	"github.com/audworth/comments-system/internal/application/post"
	"github.com/audworth/comments-system/internal/domain"
	"github.com/audworth/comments-system/internal/transport/graph/graphscalar"
)

func PostFromDomain(post *domain.Post) *Post {
	return &Post{
		ID:              post.ID.String(),
		Author:          UserFromDomain(&post.Author),
		Title:           post.Title,
		Body:            post.Body,
		CommentsEnabled: post.CommentsEnabled,
		CreatedAt:       post.CreatedAt,
		UpdatedAt:       post.UpdatedAt,
	}
}

func UserFromDomain(user *domain.User) *User {
	return &User{
		ID:   user.ID.String(),
		Name: user.Name,
	}
}

func CommentFromDomain(comment *domain.Comment) *Comment {
	var parentID *string

	if comment.ParentID != nil {
		value := comment.ParentID.String()
		parentID = &value
	}

	return &Comment{
		ID:        comment.ID.String(),
		PostID:    comment.PostID.String(),
		ParentID:  parentID,
		AuthorID:  comment.AuthorID.String(),
		Body:      comment.Body,
		CreatedAt: comment.CreatedAt,
	}
}

func PostConnectionFromPage(page *post.Page) (*PostConnection, error) {
	nodes := make([]*Post, 0, len(page.Posts))

	for _, it := range page.Posts {
		nodes = append(nodes, PostFromDomain(&it))
	}

	var endCursor *graphscalar.Cursor
	if page.Next != nil {
		enc, err := graphscalar.EncodeCursor(
			page.Next.CreatedAt,
			page.Next.ID,
		)
		if err != nil {
			return nil, fmt.Errorf("encode posts end cursor: %w", err)
		}
		endCursor = &enc
	}

	return &PostConnection{
		Nodes: nodes,
		PageInfo: &PageInfo{
			EndCursor:   endCursor,
			HasNextPage: page.HasNextPage,
		},
	}, nil
}

func CommentConnectionFromPage(page *comment.Page) (*CommentConnection, error) {
	nodes := make([]*Comment, 0, len(page.Comments))

	for _, item := range page.Comments {
		nodes = append(nodes, CommentFromDomain(&item))
	}

	var endCursor *graphscalar.Cursor
	if page.EndCursor != nil {
		enc, err := graphscalar.EncodeCursor(
			page.EndCursor.CreatedAt,
			page.EndCursor.ID,
		)
		if err != nil {
			return nil, fmt.Errorf("encode comments end cursor: %w", err)
		}
		endCursor = &enc
	}

	return &CommentConnection{
		Nodes: nodes,
		PageInfo: &PageInfo{
			EndCursor:   endCursor,
			HasNextPage: page.HasNextPage,
		},
	}, nil
}
