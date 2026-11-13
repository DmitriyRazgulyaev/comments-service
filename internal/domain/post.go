package domain

import "time"

type Post struct {
	Id              int64
	Title           string
	Content         string
	AuthorId        int64
	CommentsAllowed bool
	CreatedAt       time.Time
}
