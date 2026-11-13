package domain

import "time"

type Comment struct {
	Id              int64
	Content         string
	CreatedAt       time.Time
	ParentCommentId *int64 // ParentCommentId == nil - корневой комментарий
	PostId          int64
	AuthorId        int64
}
