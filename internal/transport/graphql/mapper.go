package graphql

import (
	"comments-service/internal/domain"
	"comments-service/internal/service"
	"comments-service/internal/transport/graphql/model"
	"strconv"
	"time"
)

func toGraphQLPost(post *domain.Post) *model.Post {
	if post == nil {
		return nil
	}

	return &model.Post{
		ID:              strconv.FormatInt(post.Id, 10),
		AuthorID:        strconv.FormatInt(post.AuthorId, 10),
		Title:           post.Title,
		Content:         post.Content,
		CommentsAllowed: post.CommentsAllowed,
		CreatedAt:       post.CreatedAt.Format(time.RFC3339Nano),
	}
}

func toPostConnection(page service.PostPage) *model.PostConnection {
	posts := make([]*model.Post, 0, len(page.Posts))

	for i := range page.Posts {
		posts = append(posts, toGraphQLPost(&page.Posts[i]))
	}

	var endCursor *string

	if page.NextCursor != nil {
		cursor := encodePostCursor(*page.NextCursor)
		endCursor = &cursor
	}

	return &model.PostConnection{
		Nodes: posts,
		PageInfo: &model.PageInfo{
			HasNextPage: page.HasNext,
			EndCursor:   endCursor,
		},
	}
}

func toGraphQLComment(comment *domain.Comment) *model.Comment {
	if comment == nil {
		return nil
	}

	var parentID *string

	if comment.ParentCommentId != nil {
		value := strconv.FormatInt(*comment.ParentCommentId, 10)
		parentID = &value
	}

	return &model.Comment{
		ID:              strconv.FormatInt(comment.Id, 10),
		PostID:          strconv.FormatInt(comment.PostId, 10),
		ParentCommentID: parentID,
		AuthorID:        strconv.FormatInt(comment.AuthorId, 10),
		Content:         comment.Content,
		CreatedAt:       comment.CreatedAt.Format(time.RFC3339Nano),
	}
}

func toCommentConnection(page service.CommentPage) *model.CommentConnection {
	comments := make([]*model.Comment, 0, len(page.Comments))

	for i := range page.Comments {
		comments = append(comments, toGraphQLComment(&page.Comments[i]))
	}

	var endCursor *string

	if page.NextCursor != nil {
		encoded := encodeCommentCursor(*page.NextCursor)
		endCursor = &encoded
	}

	return &model.CommentConnection{
		Nodes: comments,
		PageInfo: &model.PageInfo{
			HasNextPage: page.HasNext,
			EndCursor:   endCursor,
		},
	}
}
