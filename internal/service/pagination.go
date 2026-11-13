package service

import (
	"comments-service/internal/domain"
	"time"
)

type GetRepliesParams struct {
	PostId          int64
	ParentCommentId *int64 // ParentCommentId == nil - корневой комментарий
	Limit           int
	After           *CommentCursor
}

type CommentCursor struct {
	CreatedAt time.Time
	ID        int64
}

type CommentPage struct {
	Comments   []domain.Comment
	NextCursor *CommentCursor
	HasNext    bool
}

type ListPostsParams struct {
	Limit int
	After *int64
}

type PostPage struct {
	Posts      []domain.Post
	NextCursor *int64
	HasNext    bool
}
