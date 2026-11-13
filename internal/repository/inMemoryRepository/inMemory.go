package inMemoryRepository

import (
	"context"
	"sync"

	"comments-service/internal/domain"
	"comments-service/internal/service"
)

type state struct {
	posts    map[int64]domain.Post
	comments map[int64]domain.Comment

	nextPostId    int64
	nextCommentId int64
}

type UnitOfWork struct {
	mu    sync.Mutex
	state *state
}

func NewUnitOfWork() *UnitOfWork {
	return &UnitOfWork{
		state: &state{
			posts:         make(map[int64]domain.Post),
			comments:      make(map[int64]domain.Comment),
			nextPostId:    1,
			nextCommentId: 1,
		},
	}
}

type txRepositories struct {
	postRepo    *PostRepository
	commentRepo *CommentRepository
}

func (r *txRepositories) GetPostRepository() service.PostRepository {
	return r.postRepo
}

func (r *txRepositories) GetCommentRepository() service.CommentRepository {
	return r.commentRepo
}

func (u *UnitOfWork) Transaction(ctx context.Context, fn func(service.Repositories) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}

	txState := cloneState(u.state)

	repos := &txRepositories{
		postRepo: &PostRepository{
			state: txState,
		},
		commentRepo: &CommentRepository{
			state: txState,
		},
	}

	if err := fn(repos); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	u.state = txState

	return nil
}

func cloneState(src *state) *state {
	dst := &state{
		posts: make(
			map[int64]domain.Post,
			len(src.posts),
		),
		comments: make(
			map[int64]domain.Comment,
			len(src.comments),
		),
		nextPostId:    src.nextPostId,
		nextCommentId: src.nextCommentId,
	}

	for id, post := range src.posts {
		dst.posts[id] = post
	}

	for id, comment := range src.comments {
		dst.comments[id] = cloneComment(comment)
	}

	return dst
}

func cloneComment(comment domain.Comment) domain.Comment {
	if comment.ParentCommentId != nil {
		id := *comment.ParentCommentId
		comment.ParentCommentId = &id
	}

	return comment
}
