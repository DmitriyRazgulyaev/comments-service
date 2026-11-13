package service

import (
	"comments-service/internal/domain"
	"context"
)

type CommentPublisher interface {
	Publish(ctx context.Context, comment *domain.Comment) error
}

type CommentSubscriber interface {
	Subscribe(ctx context.Context, postID int64) (<-chan *domain.Comment, error)
}
