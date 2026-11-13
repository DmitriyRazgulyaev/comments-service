package graphql

import (
	"context"

	"comments-service/internal/domain"
	"comments-service/internal/service"
)

type PostService interface {
	CreatePost(ctx context.Context, input service.CreatePostInput) (*domain.Post, error)
	GetPost(ctx context.Context, id int64) (*domain.Post, error)
	ListPosts(ctx context.Context, params service.ListPostsParams) (service.PostPage, error)
	SetCommentsAllowed(ctx context.Context, input service.SetCommentsAllowedInput) error
}

type CommentService interface {
	CreateComment(ctx context.Context, input service.CreateCommentInput) (*domain.Comment, error)
	GetComments(ctx context.Context, params service.GetRepliesParams) (service.CommentPage, error)
}

type CommentSubscriber interface {
	Subscribe(ctx context.Context, postID int64) (<-chan *domain.Comment, error)
}
