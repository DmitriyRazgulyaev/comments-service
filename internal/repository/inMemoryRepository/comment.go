package inMemoryRepository

import (
	"comments-service/internal/domain"
	"comments-service/internal/service"
	"context"
	"fmt"
	"sort"
	"time"
	"unicode/utf8"
)

type CommentRepository struct {
	state *state
}

func (r *CommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if comment == nil {
		return fmt.Errorf("create comment: nil comment")
	}

	if _, exists := r.state.posts[comment.PostId]; !exists {
		return domain.ErrPostNotFound
	}

	if comment.ParentCommentId != nil {
		parent, exists := r.state.comments[*comment.ParentCommentId]

		if !exists {
			return domain.ErrCommentNotFound
		}

		if parent.PostId != comment.PostId {
			return domain.ErrInvalidParent
		}
	}

	if size := utf8.RuneCountInString(comment.Content);
		size < 1 || size > 2000 {
		return fmt.Errorf(
			"comment length must be between 1 and 2000",
		)
	}

	stored := cloneComment(*comment)

	stored.Id = r.state.nextCommentId
	stored.CreatedAt = time.Now().UTC().Truncate(
		time.Microsecond,
	)

	r.state.comments[stored.Id] = stored
	r.state.nextCommentId++

	*comment = cloneComment(stored)

	return nil
}

func (r *CommentRepository) GetById(ctx context.Context, id int64) (*domain.Comment, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	comment, exists := r.state.comments[id]
	if !exists {
		return nil, domain.ErrCommentNotFound
	}

	commentCopy := cloneComment(comment)

	return &commentCopy, nil
}

func (r *CommentRepository) GetReplies(ctx context.Context, params service.GetRepliesParams) (service.CommentPage, error) {
	if err := ctx.Err(); err != nil {
		return service.CommentPage{}, err
	}

	if params.PostId <= 0 {
		return service.CommentPage{},
			fmt.Errorf("invalid post ID")
	}

	if params.Limit < 1 || params.Limit > 100 {
		return service.CommentPage{},
			fmt.Errorf("invalid comments limit")
	}

	if params.ParentCommentId != nil && *params.ParentCommentId <= 0 {
		return service.CommentPage{},
			fmt.Errorf("invalid parent comment ID")
	}

	if params.After != nil && (params.After.ID <= 0 || params.After.CreatedAt.IsZero()) {
		return service.CommentPage{},
			fmt.Errorf("invalid comment cursor")
	}

	comments := make([]domain.Comment, 0)

	for _, comment := range r.state.comments {
		if comment.PostId != params.PostId {
			continue
		}

		if params.ParentCommentId == nil {
			if comment.ParentCommentId != nil {
				continue
			}
		} else {
			if comment.ParentCommentId == nil || *comment.ParentCommentId != *params.ParentCommentId {
				continue
			}
		}

		if params.After != nil {
			if comment.CreatedAt.Before(params.After.CreatedAt) {
				continue
			}

			if comment.CreatedAt.Equal(params.After.CreatedAt) && comment.Id <= params.After.ID {
				continue
			}
		}

		comments = append(comments, cloneComment(comment))
	}

	sort.Slice(comments, func(i, j int) bool {
		if comments[i].CreatedAt.Equal(comments[j].CreatedAt) {
			return comments[i].Id < comments[j].Id
		}

		return comments[i].CreatedAt.Before(comments[j].CreatedAt)
	})

	hasNext := len(comments) > params.Limit

	if hasNext {
		comments = comments[:params.Limit]
	}

	page := service.CommentPage{
		Comments: comments,
		HasNext:  hasNext,
	}

	if hasNext && len(comments) > 0 {
		last := comments[len(comments)-1]

		page.NextCursor = &service.CommentCursor{
			CreatedAt: last.CreatedAt,
			ID:        last.Id,
		}
	}

	return page, nil
}
