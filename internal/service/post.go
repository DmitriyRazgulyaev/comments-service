package service

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"comments-service/internal/domain"
)

func (s *CommentsService) CreatePost(ctx context.Context, input CreatePostInput) (*domain.Post, error) {
	if input.AuthorId <= 0 {
		return nil, fmt.Errorf("invalid author id")
	}

	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(input.Content)

	if title == "" {
		return nil, fmt.Errorf("post title cannot be empty")
	}

	if content == "" {
		return nil, fmt.Errorf("post content cannot be empty")
	}

	if utf8.RuneCountInString(title) > 50 {
		return nil, fmt.Errorf("post title is too long")
	}

	post := &domain.Post{
		AuthorId:        input.AuthorId,
		Title:           title,
		Content:         content,
		CommentsAllowed: input.CommentsAllowed,
	}

	err := s.unitOfWork.Transaction(
		ctx,
		func(repos Repositories) error {
			return repos.GetPostRepository().Create(ctx, post)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("createPost.Transaction: %w", err)
	}

	return post, nil
}

func (s *CommentsService) GetPost(ctx context.Context, id int64) (*domain.Post, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid post id")
	}

	var post *domain.Post

	err := s.unitOfWork.Transaction(
		ctx,
		func(repos Repositories) error {
			var err error

			post, err = repos.GetPostRepository().GetById(ctx, id)

			return err
		},
	)
	if err != nil {
		return nil, fmt.Errorf("getPost.Transaction: %w", err)
	}

	return post, nil
}

func (s *CommentsService) ListPosts(ctx context.Context, params ListPostsParams) (PostPage, error) {
	if params.Limit <= 0 || params.Limit > 100 {
		return PostPage{}, fmt.Errorf(
			"limit must be between 1 and 100",
		)
	}

	if params.After != nil && *params.After <= 0 {
		return PostPage{}, fmt.Errorf("invalid post cursor")
	}

	var page PostPage

	err := s.unitOfWork.Transaction(
		ctx,
		func(repos Repositories) error {
			var err error

			page, err = repos.GetPostRepository().List(ctx, params)

			return err
		},
	)
	if err != nil {
		return PostPage{}, fmt.Errorf("ListPosts.Transaction: %w", err)
	}

	return page, nil
}

func (s *CommentsService) SetCommentsAllowed(ctx context.Context, input SetCommentsAllowedInput) error {
	if input.PostId <= 0 {
		return fmt.Errorf("invalid post id")
	}

	if input.AuthorId <= 0 {
		return fmt.Errorf("invalid author id")
	}

	err := s.unitOfWork.Transaction(
		ctx,
		func(repos Repositories) error {
			return repos.GetPostRepository().SetCommentsAllowed(
				ctx,
				input.PostId,
				input.AuthorId,
				input.Allowed,
			)
		},
	)
	if err != nil {
		return fmt.Errorf("setCommentsAllowed.Transaction: %w", err)
	}

	return nil
}
