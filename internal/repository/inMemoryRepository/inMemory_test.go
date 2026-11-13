package inMemoryRepository

import (
	"comments-service/internal/domain"
	"comments-service/internal/service"
	"context"
	"errors"
	"testing"
)

func TestTransactionRollback(t *testing.T) {
	uow := NewUnitOfWork()
	ctx := context.Background()

	expectedErr := errors.New("rollback requested")

	err := uow.Transaction(
		ctx,
		func(repos service.Repositories) error {
			post := &domain.Post{
				AuthorId:        1,
				Title:           "Test post",
				Content:         "Test content",
				CommentsAllowed: true,
			}

			if err := repos.GetPostRepository().Create(ctx, post); err != nil {
				return err
			}

			return expectedErr
		},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected rollback error, got: %v", err)
	}

	err = uow.Transaction(
		ctx,
		func(repos service.Repositories) error {
			page, err := repos.GetPostRepository().List(ctx, service.ListPostsParams{Limit: 10})
			if err != nil {
				return err
			}

			if len(page.Posts) != 0 {
				t.Fatalf("rollback failed: found %d posts", len(page.Posts))
			}

			return nil
		},
	)

	if err != nil {
		t.Fatal(err)
	}
}

func TestTransactionCommit(t *testing.T) {
	uow := NewUnitOfWork()
	ctx := context.Background()

	var createdID int64

	err := uow.Transaction(
		ctx,
		func(repos service.Repositories) error {
			post := &domain.Post{
				AuthorId:        1,
				Title:           "Test post",
				Content:         "Test content",
				CommentsAllowed: true,
			}

			if err := repos.GetPostRepository().Create(ctx, post); err != nil {
				return err
			}

			createdID = post.Id
			return nil
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	err = uow.Transaction(
		ctx,
		func(repos service.Repositories) error {
			post, err := repos.GetPostRepository().GetById(ctx, createdID)
			if err != nil {
				return err
			}

			if post.Title != "Test post" {
				t.Fatalf("unexpected title: %s", post.Title)
			}

			return nil
		},
	)

	if err != nil {
		t.Fatal(err)
	}
}
