package domain

import "errors"

var (
	ErrPostNotFound     = errors.New("post not found")
	ErrCommentNotFound  = errors.New("comment not found")
	ErrCommentsDisabled = errors.New("comments disabled")
	ErrInvalidParent    = errors.New("invalid parent comment")
	ErrCommentTooLong   = errors.New("comment too long")
	ErrEmptyComment     = errors.New("empty comment")
	ErrForbidden        = errors.New("forbidden")
	ErrInvalidPostID    = errors.New("invalid post id")
	ErrInvalidLimit     = errors.New("invalid limit")
)
