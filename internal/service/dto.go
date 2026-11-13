package service

type CreatePostInput struct {
	AuthorId        int64
	Title           string
	Content         string
	CommentsAllowed bool
}

type CreateCommentInput struct {
	PostId          int64
	ParentCommentId *int64
	AuthorId        int64
	Content         string
}

type SetCommentsAllowedInput struct {
	PostId   int64
	AuthorId int64
	Allowed  bool
}
