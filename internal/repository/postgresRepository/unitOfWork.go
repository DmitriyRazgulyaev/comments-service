package postgresRepository

import (
	"comments-service/internal/repository/postgresRepository/commentRepository"
	"comments-service/internal/repository/postgresRepository/postRepository"
	"comments-service/internal/service"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type txRepositories struct {
	commentRepo *commentRepository.CommentRepository
	postRepo    *postRepository.PostRepository
}

func (txr *txRepositories) GetCommentRepository() service.CommentRepository {
	return txr.commentRepo
}

func (txr *txRepositories) GetPostRepository() service.PostRepository {
	return txr.postRepo
}

type UnitOfWork struct {
	pool *pgxpool.Pool
}

func NewPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("newPool.New: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("newPool.Ping: %w", err)
	}

	return pool, nil
}

func NewUnitOfWork(ctx context.Context, pool *pgxpool.Pool) *UnitOfWork {
	return &UnitOfWork{
		pool: pool,
	}
}

func (uow UnitOfWork) Transaction(ctx context.Context, fn func(repositories service.Repositories) error) error {
	tx, err := uow.pool.Begin(ctx)
	if err != nil {
		//TODO: Поставить правильную проверку ошибки
		return err
	}

	defer tx.Rollback(ctx)

	repos := &txRepositories{
		postRepo:    postRepository.New(tx),
		commentRepo: commentRepository.New(tx),
	}

	if err := fn(repos); err != nil {
		//TODO: Сделать проверку ошибки
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("transaction.Commit: %w", err)
	}

	return nil
}
