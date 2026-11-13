package graphql

import (
	"comments-service/internal/service"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
)

func encodePostCursor(id int64) string {
	return base64.RawURLEncoding.EncodeToString(
		[]byte(strconv.FormatInt(id, 10)),
	)
}

func decodePostCursor(value *string) (*int64, error) {
	if value == nil {
		return nil, nil
	}

	data, err := base64.RawURLEncoding.DecodeString(*value)
	if err != nil {
		return nil, fmt.Errorf("decode post cursor: %w", err)
	}

	id, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil || id <= 0 {
		return nil, fmt.Errorf("invalid post cursor")
	}

	return &id, nil
}

func encodeCommentCursor(cursor service.CommentCursor) string {
	data, _ := json.Marshal(cursor)

	return base64.RawURLEncoding.EncodeToString(data)
}

func decodeCommentCursor(value *string) (*service.CommentCursor, error) {
	if value == nil {
		return nil, nil
	}

	data, err := base64.RawURLEncoding.DecodeString(*value)
	if err != nil {
		return nil, fmt.Errorf("invalid comment cursor: %w", err)
	}

	var cursor service.CommentCursor

	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, fmt.Errorf("invalid comment cursor: %w", err)
	}

	if cursor.ID <= 0 || cursor.CreatedAt.IsZero() {
		return nil, fmt.Errorf(
			"invalid comment cursor",
		)
	}

	return &cursor, nil
}

func parseId(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)

	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid ID: %q", value)
	}

	return id, nil
}
