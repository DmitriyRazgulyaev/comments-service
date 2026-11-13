package postRepository

import (
	"comments-service/internal/domain"
	"comments-service/internal/service"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type PostRepository struct {
	tx pgx.Tx
}

func New(tx pgx.Tx) *PostRepository {
	return &PostRepository{
		tx: tx,
	}
}

func (pr *PostRepository) Create(ctx context.Context, post *domain.Post) error {
	query := `
        INSERT INTO posts (
        	author_id,
            title,
            content,
            comments_allowed
        )
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at
    `

	err := pr.tx.QueryRow(
		ctx,
		query,
		post.AuthorId,
		post.Title,
		post.Content,
		post.CommentsAllowed,
	).Scan(
		&post.Id,
		&post.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("create.QueryRow: %w", err)
	}

	return nil
}

func (pr *PostRepository) GetById(ctx context.Context, id int64) (*domain.Post, error) {
	query := `
		SELECT
			id,
			author_id,
			title,
			content,
			comments_allowed,
			created_at
		FROM posts
		WHERE id = $1
	`

	var post domain.Post

	err := pr.tx.QueryRow(ctx, query, id).Scan(
		&post.Id,
		&post.AuthorId,
		&post.Title,
		&post.Content,
		&post.CommentsAllowed,
		&post.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrPostNotFound
	}

	if err != nil {
		return nil, fmt.Errorf(
			"getByID.QueryRow with id = %d: %w",
			id,
			err,
		)
	}

	return &post, nil
}

func (pr *PostRepository) GetByIDForShare(ctx context.Context, id int64) (*domain.Post, error) {
	query := `
        SELECT
            id,
            author_id,
            title,
            content,
            comments_allowed,
            created_at
        FROM posts
        WHERE id = $1
        FOR SHARE
    `

	var post domain.Post

	err := pr.tx.QueryRow(ctx, query, id).Scan(
		&post.Id,
		&post.AuthorId,
		&post.Title,
		&post.Content,
		&post.CommentsAllowed,
		&post.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrPostNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("GetByIDForShare.QueryRow with id = %d: %w", id, err)
	}

	return &post, nil
}

func (pr *PostRepository) List(ctx context.Context, params service.ListPostsParams) (service.PostPage, error) {
	if params.Limit <= 0 || params.Limit > 100 {
		return service.PostPage{}, fmt.Errorf("invalid posts limit: %d", params.Limit)
	}
	query := `
    SELECT id, author_id, title, content,
           comments_allowed, created_at
    FROM posts
    ORDER BY id ASC
    LIMIT $1
`

	args := []any{params.Limit + 1}

	if params.After != nil {
		query = `
        SELECT id, author_id, title, content, comments_allowed, created_at
        FROM posts
        WHERE id > $1
        ORDER BY id ASC
        LIMIT $2
    `

		args = []any{
			*params.After,
			params.Limit + 1,
		}
	}
	rows, err := pr.tx.Query(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return service.PostPage{}, fmt.Errorf(
			"list.Query: %w",
			err,
		)
	}
	defer rows.Close()

	posts := make([]domain.Post, 0, params.Limit+1)

	for rows.Next() {
		var post domain.Post

		err := rows.Scan(
			&post.Id,
			&post.AuthorId,
			&post.Title,
			&post.Content,
			&post.CommentsAllowed,
			&post.CreatedAt,
		)
		if err != nil {
			return service.PostPage{}, fmt.Errorf("list.Scan: %w", err)
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return service.PostPage{}, fmt.Errorf("list.Iteration: %w", err)
	}

	hasNext := len(posts) > params.Limit

	if hasNext {
		posts = posts[:params.Limit]
	}

	page := service.PostPage{
		Posts:   posts,
		HasNext: hasNext,
	}

	if hasNext {
		cursor := posts[len(posts)-1].Id
		page.NextCursor = &cursor
	}

	return page, nil
}

func (pr *PostRepository) SetCommentsAllowed(ctx context.Context, postId int64, authorId int64, allowed bool) error {
	query := `
		UPDATE posts
		SET comments_allowed = $1
		WHERE id = $2 AND author_id = $3
		RETURNING id
	`

	var id int64

	err := pr.tx.QueryRow(
		ctx,
		query,
		allowed,
		postId,
		authorId,
	).Scan(&id)

	if errors.Is(err, pgx.ErrNoRows) {
		_, getErr := pr.GetById(ctx, postId)

		if errors.Is(getErr, domain.ErrPostNotFound) {
			return domain.ErrPostNotFound
		}

		if getErr != nil {
			return getErr
		}

		return domain.ErrForbidden
	}

	if err != nil {
		return fmt.Errorf("setCommentsAllowed.QueryRow with postId = %d: %w", postId, err)
	}

	return nil
}
