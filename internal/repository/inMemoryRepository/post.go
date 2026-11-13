package inMemoryRepository

import (
	"comments-service/internal/domain"
	"comments-service/internal/service"
	"context"
	"fmt"
	"slices"
	"time"
)

type PostRepository struct {
	state *state
}

func (r *PostRepository) Create(ctx context.Context, post *domain.Post) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if post == nil {
		return fmt.Errorf("create post: nil post")
	}

	stored := *post

	stored.Id = r.state.nextPostId
	stored.CreatedAt = time.Now().UTC().Truncate(
		time.Microsecond,
	)

	r.state.posts[stored.Id] = stored
	r.state.nextPostId++

	*post = stored

	return nil
}

func (r *PostRepository) GetById(ctx context.Context, id int64) (*domain.Post, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	post, exists := r.state.posts[id]
	if !exists {
		return nil, domain.ErrPostNotFound
	}

	return &post, nil
}

func (r *PostRepository) GetByIDForShare(ctx context.Context, id int64) (*domain.Post, error) {
	return r.GetById(ctx, id)
}

func (r *PostRepository) List(ctx context.Context, params service.ListPostsParams) (service.PostPage, error) {
	if err := ctx.Err(); err != nil {
		return service.PostPage{}, err
	}

	if params.Limit < 1 || params.Limit > 100 {
		return service.PostPage{},
			fmt.Errorf("invalid posts limit")
	}

	var afterID int64

	if params.After != nil {
		if *params.After <= 0 {
			return service.PostPage{}, fmt.Errorf("invalid posts cursor")
		}
		afterID = *params.After
	}

	ids := make([]int64, 0, len(r.state.posts))

	for id := range r.state.posts {
		if id > afterID {
			ids = append(ids, id)
		}
	}

	slices.Sort(ids)

	hasNext := len(ids) > params.Limit

	if hasNext {
		ids = ids[:params.Limit]
	}

	posts := make([]domain.Post, 0, len(ids))

	for _, id := range ids {
		posts = append(posts, r.state.posts[id])
	}

	page := service.PostPage{
		Posts:   posts,
		HasNext: hasNext,
	}

	if hasNext && len(posts) > 0 {
		lastID := posts[len(posts)-1].Id
		page.NextCursor = &lastID
	}

	return page, nil
}

func (r *PostRepository) SetCommentsAllowed(ctx context.Context, postID int64, authorID int64, allowed bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	post, exists := r.state.posts[postID]
	if !exists {
		return domain.ErrPostNotFound
	}

	if post.AuthorId != authorID {
		return domain.ErrForbidden
	}

	post.CommentsAllowed = allowed
	r.state.posts[postID] = post

	return nil
}
