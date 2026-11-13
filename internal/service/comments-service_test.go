package service_test

import (
	"comments-service/internal/repository/inMemoryRepository"
	"comments-service/internal/service"
	"comments-service/internal/subscription"
	"context"
	"testing"
)

func newTestService() *service.CommentsService {
	uow := inMemoryRepository.NewUnitOfWork()
	broker := subscription.NewBroker()

	return service.NewCommentsService(uow, broker)
}

func TestCreatePost(t *testing.T) {
	ctx := context.Background()
	s := newTestService()

	post, err := s.CreatePost(ctx, service.CreatePostInput{
		AuthorId:        1,
		Title:           "Test post",
		Content:         "Test content",
		CommentsAllowed: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if post.Id == 0 {
		t.Fatal("post ID was not generated")
	}

	got, err := s.GetPost(ctx, post.Id)
	if err != nil {
		t.Fatal(err)
	}

	if got.Title != "Test post" {
		t.Errorf("unexpected title: %s", got.Title)
	}
}

func TestCreateComment(t *testing.T) {
	ctx := context.Background()
	s := newTestService()

	post, err := s.CreatePost(ctx, service.CreatePostInput{
		AuthorId:        1,
		Title:           "Post",
		Content:         "Content",
		CommentsAllowed: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	comment, err := s.CreateComment(ctx, service.CreateCommentInput{
		PostId:   post.Id,
		AuthorId: 2,
		Content:  "First comment",
	})
	if err != nil {
		t.Fatal(err)
	}

	if comment.Id == 0 {
		t.Fatal("comment ID was not generated")
	}

	if comment.ParentCommentId != nil {
		t.Error("expected root comment")
	}

	page, err := s.GetComments(ctx, service.GetRepliesParams{
		PostId: post.Id,
		Limit:  10,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(page.Comments) != 1 {
		t.Fatalf(
			"expected 1 comment, got %d",
			len(page.Comments),
		)
	}
}

func TestCreateReply(t *testing.T) {
	ctx := context.Background()
	s := newTestService()

	post, err := s.CreatePost(ctx, service.CreatePostInput{
		AuthorId:        1,
		Title:           "Post",
		Content:         "Content",
		CommentsAllowed: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	parent, err := s.CreateComment(ctx, service.CreateCommentInput{
		PostId:   post.Id,
		AuthorId: 2,
		Content:  "Parent",
	})
	if err != nil {
		t.Fatal(err)
	}

	parentID := parent.Id

	reply, err := s.CreateComment(ctx, service.CreateCommentInput{
		PostId:          post.Id,
		ParentCommentId: &parentID,
		AuthorId:        3,
		Content:         "Reply",
	})
	if err != nil {
		t.Fatal(err)
	}

	if reply.ParentCommentId == nil {
		t.Fatal("reply has no parent")
	}

	if *reply.ParentCommentId != parentID {
		t.Errorf(
			"expected parent ID %d, got %d",
			parentID,
			*reply.ParentCommentId,
		)
	}

	page, err := s.GetComments(ctx, service.GetRepliesParams{
		PostId:          post.Id,
		ParentCommentId: &parentID,
		Limit:           10,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(page.Comments) != 1 {
		t.Fatalf(
			"expected 1 reply, got %d",
			len(page.Comments),
		)
	}

	if page.Comments[0].Id != reply.Id {
		t.Error("unexpected reply returned")
	}
}

func TestCommentsDisabled(t *testing.T) {
	ctx := context.Background()
	s := newTestService()

	post, err := s.CreatePost(ctx, service.CreatePostInput{
		AuthorId:        1,
		Title:           "Post",
		Content:         "Content",
		CommentsAllowed: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = s.SetCommentsAllowed(ctx, service.SetCommentsAllowedInput{
		PostId:   post.Id,
		AuthorId: 1,
		Allowed:  false,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.CreateComment(ctx, service.CreateCommentInput{
		PostId:   post.Id,
		AuthorId: 2,
		Content:  "Comment",
	})

	if err == nil {
		t.Fatal("expected error when comments are disabled")
	}

	page, err := s.GetComments(ctx, service.GetRepliesParams{
		PostId: post.Id,
		Limit:  10,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(page.Comments) != 0 {
		t.Error("comment was saved despite disabled comments")
	}
}

func TestSetCommentsAllowedWrongAuthor(t *testing.T) {
	ctx := context.Background()
	s := newTestService()

	post, err := s.CreatePost(ctx, service.CreatePostInput{
		AuthorId:        1,
		Title:           "Post",
		Content:         "Content",
		CommentsAllowed: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = s.SetCommentsAllowed(ctx, service.SetCommentsAllowedInput{
		PostId:   post.Id,
		AuthorId: 999,
		Allowed:  false,
	})

	if err == nil {
		t.Fatal("expected error for wrong author")
	}

	got, err := s.GetPost(ctx, post.Id)
	if err != nil {
		t.Fatal(err)
	}

	if !got.CommentsAllowed {
		t.Error("comments were disabled by wrong author")
	}
}
