package subscription_test

import (
	"context"
	"testing"
	"time"

	"comments-service/internal/domain"
	"comments-service/internal/subscription"
)

func TestBrokerPublish(t *testing.T) {
	broker := subscription.NewBroker()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := broker.Subscribe(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	comment := &domain.Comment{
		Id:      10,
		PostId:  1,
		Content: "Test comment",
	}

	err = broker.Publish(ctx, comment)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case received := <-ch:
		if received == nil {
			t.Fatal("received nil comment")
		}

		if received.Id != comment.Id {
			t.Errorf(
				"expected comment ID %d, got %d",
				comment.Id,
				received.Id,
			)
		}

	case <-time.After(time.Second):
		t.Fatal("comment was not delivered")
	}
}

func TestBrokerDifferentPosts(t *testing.T) {
	broker := subscription.NewBroker()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch1, err := broker.Subscribe(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	ch2, err := broker.Subscribe(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}

	comment := &domain.Comment{
		Id:      5,
		PostId:  1,
		Content: "Comment for post 1",
	}

	if err := broker.Publish(ctx, comment); err != nil {
		t.Fatal(err)
	}

	select {
	case received := <-ch1:
		if received == nil || received.Id != comment.Id {
			t.Fatal("unexpected comment received")
		}

	case <-time.After(time.Second):
		t.Fatal("post 1 subscriber received nothing")
	}

	select {
	case received := <-ch2:
		t.Fatalf("unexpected comment for post 2: %v", received)

	default:
	}
}

func TestBrokerUnsubscribe(t *testing.T) {
	broker := subscription.NewBroker()

	ctx, cancel := context.WithCancel(context.Background())

	ch, err := broker.Subscribe(ctx, 1)
	if err != nil {
		cancel()
		t.Fatal(err)
	}

	cancel()

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("subscription channel is still open")
		}

	case <-time.After(time.Second):
		t.Fatal("subscription channel was not closed")
	}
}
