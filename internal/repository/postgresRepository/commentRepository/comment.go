package commentRepository

import (
	"comments-service/internal/domain"
	"comments-service/internal/service"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type CommentRepository struct {
	tx pgx.Tx
}

func New(tx pgx.Tx) *CommentRepository {
	return &CommentRepository{
		tx: tx,
	}
}

func (cr *CommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	query := `
        INSERT INTO comments (
        	post_id,
            parent_comment_id,
            author_id,
            content
        )
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at
    `

	err := cr.tx.QueryRow(
		ctx,
		query,
		comment.PostId,
		comment.ParentCommentId,
		comment.AuthorId,
		comment.Content,
	).Scan(
		&comment.Id,
		&comment.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("create.QueryRow: %w", err)
	}

	return nil
}

func (cr *CommentRepository) GetById(ctx context.Context, id int64) (*domain.Comment, error) {
	query := `
        SELECT
            id,
            post_id,
            parent_comment_id,
            author_id,
            content,
            created_at
        FROM comments
        WHERE id = $1
    `

	var comment domain.Comment

	err := cr.tx.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&comment.Id,
		&comment.PostId,
		&comment.ParentCommentId,
		&comment.AuthorId,
		&comment.Content,
		&comment.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrCommentNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("getByid.QueryRow with id = %d: %w", id, err)
	}

	return &comment, nil
}

func (cr *CommentRepository) GetReplies(ctx context.Context, params service.GetRepliesParams) (service.CommentPage, error) {
	if params.Limit <= 0 || params.Limit > 100 {
		return service.CommentPage{}, fmt.Errorf(
			"invalid pagination limit: %d", params.Limit,
		)
	}

	query := `
    SELECT
        id,
        post_id,
        parent_comment_id,
        author_id,
        content,
        created_at
    FROM comments
    WHERE post_id = $1
`

	args := []any{params.PostId}

	if params.ParentCommentId == nil {
		query += `
        AND parent_comment_id IS NULL
    `
	} else {
		query += `
        AND parent_comment_id = $2
    `
		args = append(args, *params.ParentCommentId)
	}

	if params.After != nil {
		cursorTimeParam := len(args) + 1
		cursorIDParam := len(args) + 2

		query += fmt.Sprintf(`
        AND (created_at, id) > ($%d, $%d)
    `, cursorTimeParam, cursorIDParam)

		args = append(
			args,
			params.After.CreatedAt,
			params.After.ID,
		)
	}

	limitParam := len(args) + 1

	query += fmt.Sprintf(`
    ORDER BY created_at ASC, id ASC
    LIMIT $%d
	`, limitParam)

	args = append(args, params.Limit+1)

	rows, err := cr.tx.Query(ctx, query, args...)
	if err != nil {
		return service.CommentPage{}, fmt.Errorf(
			"getReplies.Query: %w", err,
		)
	}
	defer rows.Close()

	comments := make([]domain.Comment, 0, params.Limit+1)
	for rows.Next() {
		var comment domain.Comment

		err := rows.Scan(
			&comment.Id,
			&comment.PostId,
			&comment.ParentCommentId,
			&comment.AuthorId,
			&comment.Content,
			&comment.CreatedAt,
		)
		if err != nil {
			return service.CommentPage{}, fmt.Errorf(
				"getReplies.Scan: %w", err,
			)
		}

		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return service.CommentPage{}, fmt.Errorf("getReplies.Iteration: %w", err)
	}

	hasNext := len(comments) > params.Limit
	if hasNext {
		comments = comments[:params.Limit]
	}

	page := service.CommentPage{
		Comments: comments,
		HasNext:  hasNext,
	}

	if hasNext {
		last := comments[len(comments)-1]

		page.NextCursor = &service.CommentCursor{
			CreatedAt: last.CreatedAt,
			ID:        last.Id,
		}
	}

	return page, nil
}
