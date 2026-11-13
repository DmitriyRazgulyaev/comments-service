package postgresRepository_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"comments-service/internal/domain"
	postgresRepository "comments-service/internal/repository/postgresRepository"
	"comments-service/internal/service"
)

func TestTransactionCommit(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()

	pool, err := postgresRepository.NewPool(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(pool.Close)

	uow := postgresRepository.NewUnitOfWork(ctx, pool)

	post := &domain.Post{
		AuthorId:        1,
		Title:           "Commit test",
		Content:         "Content",
		CommentsAllowed: true,
	}

	err = uow.Transaction(ctx, func(repos service.Repositories) error {
		return repos.GetPostRepository().Create(ctx, post)
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, err := pool.Exec(
			context.Background(),
			"DELETE FROM posts WHERE id = $1",
			post.Id,
		)
		if err != nil {
			t.Error(err)
		}
	})

	var exists bool

	err = pool.QueryRow(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)",
		post.Id,
	).Scan(&exists)
	if err != nil {
		t.Fatal(err)
	}

	if !exists {
		t.Fatal("post was not committed")
	}
}

func TestTransactionRollback(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()

	pool, err := postgresRepository.NewPool(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	uow := postgresRepository.NewUnitOfWork(ctx, pool)

	expectedErr := errors.New("rollback")
	var postID int64

	err = uow.Transaction(ctx, func(repos service.Repositories) error {
		post := &domain.Post{
			AuthorId:        1,
			Title:           "Rollback test",
			Content:         "Content",
			CommentsAllowed: true,
		}

		if err := repos.GetPostRepository().Create(ctx, post); err != nil {
			return err
		}

		postID = post.Id
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected rollback error, got %v", err)
	}

	var exists bool

	err = pool.QueryRow(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)",
		postID,
	).Scan(&exists)
	if err != nil {
		t.Fatal(err)
	}

	if exists {
		t.Fatal("post was saved after rollback")
	}
}
