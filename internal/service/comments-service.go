package service

import (
	"comments-service/internal/domain"
	"context"
	"log"
	"strings"
	"unicode/utf8"
)

type Repositories interface {
	GetPostRepository() PostRepository
	GetCommentRepository() CommentRepository
}

type UnitOfWork interface {
	Transaction(ctx context.Context, fn func(repositories Repositories) error) error
}

type CommentRepository interface {
	Create(ctx context.Context, comment *domain.Comment) error
	GetById(ctx context.Context, id int64) (*domain.Comment, error)
	GetReplies(ctx context.Context, params GetRepliesParams) (CommentPage, error)
}

type PostRepository interface {
	Create(ctx context.Context, post *domain.Post) error
	GetById(ctx context.Context, id int64) (*domain.Post, error)
	GetByIDForShare(ctx context.Context, id int64) (*domain.Post, error)
	List(ctx context.Context, params ListPostsParams) (PostPage, error)
	SetCommentsAllowed(ctx context.Context, postId int64, authorId int64, allowed bool) error
}

type CommentsService struct {
	unitOfWork UnitOfWork
	publisher  CommentPublisher
}

func NewCommentsService(uow UnitOfWork, publisher CommentPublisher) *CommentsService {
	return &CommentsService{
		unitOfWork: uow,
		publisher:  publisher,
	}
}

func (s *CommentsService) CreateComment(ctx context.Context, input CreateCommentInput) (*domain.Comment, error) {
	content := strings.TrimSpace(input.Content)

	if content == "" {
		return nil, domain.ErrEmptyComment
	}

	if utf8.RuneCountInString(content) > 2000 {
		return nil, domain.ErrCommentTooLong
	}

	comment := &domain.Comment{
		PostId:          input.PostId,
		ParentCommentId: input.ParentCommentId,
		AuthorId:        input.AuthorId,
		Content:         content,
	}

	err := s.unitOfWork.Transaction(
		ctx,
		func(repos Repositories) error {
			post, err := repos.GetPostRepository().GetByIDForShare(ctx, input.PostId)
			if err != nil {
				return err
			}

			if !post.CommentsAllowed {
				return domain.ErrCommentsDisabled
			}

			if input.ParentCommentId != nil {
				parent, err := repos.GetCommentRepository().GetById(ctx, *input.ParentCommentId)
				if err != nil {
					return err
				}

				if parent.PostId != input.PostId {
					return domain.ErrInvalidParent
				}
			}

			return repos.GetCommentRepository().Create(ctx, comment)
		},
	)

	if err != nil {
		return nil, err
	}

	if err := s.publisher.Publish(context.Background(), comment); err != nil {
		log.Printf("createComment.Publish: %s", err)
	}

	return comment, nil
}

func (s *CommentsService) GetComments(ctx context.Context, params GetRepliesParams) (CommentPage, error) {
	var page CommentPage

	if params.PostId <= 0 {
		return page, domain.ErrInvalidPostID
	}

	if params.Limit <= 0 || params.Limit > 100 {
		return page, domain.ErrInvalidLimit
	}

	if params.ParentCommentId != nil && *params.ParentCommentId <= 0 {
		return page, domain.ErrInvalidParent
	}

	err := s.unitOfWork.Transaction(ctx, func(repos Repositories) error {
		var err error

		page, err = repos.GetCommentRepository().GetReplies(ctx, params)

		return err
	},
	)

	if err != nil {
		return CommentPage{}, err
	}

	return page, nil
}
