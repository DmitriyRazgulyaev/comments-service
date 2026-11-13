package graphql

import (
	"comments-service/internal/service"
	"testing"
	"time"
)

func TestParseID(t *testing.T) {
	id, err := parseId("123")
	if err != nil {
		t.Fatal(err)
	}

	if id != 123 {
		t.Errorf("expected 123, got %d", id)
	}
}

func TestParseIDInvalid(t *testing.T) {
	_, err := parseId("abc")

	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
}

func TestPostCursor(t *testing.T) {
	originalID := int64(42)

	encoded := encodePostCursor(originalID)

	decoded, err := decodePostCursor(&encoded)
	if err != nil {
		t.Fatal(err)
	}

	if decoded == nil {
		t.Fatal("decoded cursor is nil")
	}

	if *decoded != originalID {
		t.Errorf(
			"expected ID %d, got %d",
			originalID,
			*decoded,
		)
	}
}

func TestCommentCursor(t *testing.T) {
	original := service.CommentCursor{
		ID:        15,
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}

	encoded := encodeCommentCursor(original)

	decoded, err := decodeCommentCursor(&encoded)
	if err != nil {
		t.Fatal(err)
	}

	if decoded == nil {
		t.Fatal("decoded cursor is nil")
	}

	if decoded.ID != original.ID {
		t.Errorf(
			"expected ID %d, got %d",
			original.ID,
			decoded.ID,
		)
	}

	if !decoded.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf(
			"expected time %v, got %v",
			original.CreatedAt,
			decoded.CreatedAt,
		)
	}
}

// Проверяем некорректный курсор.
func TestInvalidCursors(t *testing.T) {
	invalid := "not-a-valid-cursor!"

	_, err := decodePostCursor(&invalid)
	if err == nil {
		t.Error("expected error for invalid post cursor")
	}

	_, err = decodeCommentCursor(&invalid)
	if err == nil {
		t.Error("expected error for invalid comment cursor")
	}
}

func TestNilCursor(t *testing.T) {
	postCursor, err := decodePostCursor(nil)
	if err != nil {
		t.Fatal(err)
	}

	if postCursor != nil {
		t.Error("expected nil post cursor")
	}

	commentCursor, err := decodeCommentCursor(nil)
	if err != nil {
		t.Fatal(err)
	}

	if commentCursor != nil {
		t.Error("expected nil comment cursor")
	}
}
