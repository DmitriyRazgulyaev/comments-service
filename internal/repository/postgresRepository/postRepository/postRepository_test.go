package postRepository_test

import (
	"context"
	"os"
	"testing"

	"comments-service/internal/domain"
	postgresRepository "comments-service/internal/repository/postgresRepository"
	"comments-service/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	pool, err := postgresRepository.NewPool(
		context.Background(),
		dsn,
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(pool.Close)
	return pool
}

func TestCreateAndGetPost(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	uow := postgresRepository.NewUnitOfWork(ctx, pool)

	post := &domain.Post{
		Title:           "Test post",
		Content:         "Test content",
		AuthorId:        1,
		CommentsAllowed: true,
	}

	err := uow.Transaction(ctx, func(repos service.Repositories) error {
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

	var got *domain.Post

	err = uow.Transaction(ctx, func(repos service.Repositories) error {
		var err error
		got, err = repos.GetPostRepository().GetById(ctx, post.Id)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	if got.Id != post.Id {
		t.Errorf("expected ID %d, got %d", post.Id, got.Id)
	}

	if got.Title != post.Title {
		t.Errorf("expected title %q, got %q", post.Title, got.Title)
	}
}

func TestSetCommentsAllowed(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	uow := postgresRepository.NewUnitOfWork(ctx, pool)

	post := &domain.Post{
		Title:           "Test post",
		Content:         "Content",
		AuthorId:        1,
		CommentsAllowed: true,
	}

	err := uow.Transaction(ctx, func(repos service.Repositories) error {
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

	err = uow.Transaction(ctx, func(repos service.Repositories) error {
		return repos.GetPostRepository().SetCommentsAllowed(
			ctx,
			post.Id,
			post.AuthorId,
			false,
		)
	})
	if err != nil {
		t.Fatal(err)
	}

	var got *domain.Post

	err = uow.Transaction(ctx, func(repos service.Repositories) error {
		var err error
		got, err = repos.GetPostRepository().GetById(ctx, post.Id)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	if got.CommentsAllowed {
		t.Fatal("comments should be disabled")
	}

	// Другой автор не должен изменить настройку.
	err = uow.Transaction(ctx, func(repos service.Repositories) error {
		return repos.GetPostRepository().SetCommentsAllowed(
			ctx,
			post.Id,
			999,
			true,
		)
	})

	if err == nil {
		t.Fatal("expected error for wrong author")
	}
}
