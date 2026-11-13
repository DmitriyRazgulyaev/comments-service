package commentRepository_test

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

func TestCreateAndGetReplies(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	uow := postgresRepository.NewUnitOfWork(ctx, pool)

	post := &domain.Post{
		Title:           "Post",
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

	parent := &domain.Comment{
		PostId:   post.Id,
		AuthorId: 2,
		Content:  "Parent",
	}

	err = uow.Transaction(ctx, func(repos service.Repositories) error {
		if err := repos.GetCommentRepository().Create(ctx, parent); err != nil {
			return err
		}

		reply := &domain.Comment{
			PostId:          post.Id,
			ParentCommentId: &parent.Id,
			AuthorId:        3,
			Content:         "Reply",
		}

		return repos.GetCommentRepository().Create(ctx, reply)
	})
	if err != nil {
		t.Fatal(err)
	}

	var roots service.CommentPage
	var replies service.CommentPage

	err = uow.Transaction(ctx, func(repos service.Repositories) error {
		var err error

		roots, err = repos.GetCommentRepository().GetReplies(
			ctx,
			service.GetRepliesParams{
				PostId: post.Id,
				Limit:  10,
			},
		)
		if err != nil {
			return err
		}

		replies, err = repos.GetCommentRepository().GetReplies(
			ctx,
			service.GetRepliesParams{
				PostId:          post.Id,
				ParentCommentId: &parent.Id,
				Limit:           10,
			},
		)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(roots.Comments) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots.Comments))
	}

	if len(replies.Comments) != 1 {
		t.Fatalf("expected 1 reply, got %d", len(replies.Comments))
	}

	if replies.Comments[0].Content != "Reply" {
		t.Errorf(
			"unexpected reply content: %s",
			replies.Comments[0].Content,
		)
	}
}

func TestInvalidParentPost(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	uow := postgresRepository.NewUnitOfWork(ctx, pool)

	post1 := &domain.Post{
		Title:           "Post 1",
		Content:         "Content",
		AuthorId:        1,
		CommentsAllowed: true,
	}

	post2 := &domain.Post{
		Title:           "Post 2",
		Content:         "Content",
		AuthorId:        1,
		CommentsAllowed: true,
	}

	err := uow.Transaction(ctx, func(repos service.Repositories) error {
		if err := repos.GetPostRepository().Create(ctx, post1); err != nil {
			return err
		}
		return repos.GetPostRepository().Create(ctx, post2)
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		for _, id := range []int64{post1.Id, post2.Id} {
			_, err := pool.Exec(
				context.Background(),
				"DELETE FROM posts WHERE id = $1",
				id,
			)
			if err != nil {
				t.Error(err)
			}
		}
	})

	parent := &domain.Comment{
		PostId:   post1.Id,
		AuthorId: 2,
		Content:  "Parent",
	}

	err = uow.Transaction(ctx, func(repos service.Repositories) error {
		return repos.GetCommentRepository().Create(ctx, parent)
	})
	if err != nil {
		t.Fatal(err)
	}

	// Пытаемся добавить ответ в другой пост.
	invalidReply := &domain.Comment{
		PostId:          post2.Id,
		ParentCommentId: &parent.Id,
		AuthorId:        3,
		Content:         "Invalid reply",
	}

	err = uow.Transaction(ctx, func(repos service.Repositories) error {
		return repos.GetCommentRepository().Create(ctx, invalidReply)
	})

	if err == nil {
		t.Fatal("expected error for parent from another post")
	}
}
