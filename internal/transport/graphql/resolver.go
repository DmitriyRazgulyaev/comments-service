package graphql

type Resolver struct {
	posts      PostService
	comments   CommentService
	subscriber CommentSubscriber
}

func NewResolver(posts PostService, comments CommentService, subscriber CommentSubscriber) *Resolver {
	return &Resolver{
		posts:      posts,
		comments:   comments,
		subscriber: subscriber,
	}
}
