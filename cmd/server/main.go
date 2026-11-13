package main

import (
	"comments-service/internal/config"
	"comments-service/internal/repository/inMemoryRepository"
	"comments-service/internal/repository/postgresRepository"
	"comments-service/internal/service"
	"comments-service/internal/subscription"
	"comments-service/internal/transport/graphql"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	var envPath string
	if envPath = os.Getenv("ENV_PATH"); envPath == "" {
		envPath = "./config/.env"
	}

	cfg, err := config.ParseConfigFromEnv(envPath)
	if err != nil {
		panic(err)
	}

	var uow service.UnitOfWork
	switch cfg.SrvConfig.Storage {
	case "memory":
		uow = inMemoryRepository.NewUnitOfWork()
	case "postgres":
		dataSource := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable",
			cfg.DBConfig.User, cfg.DBConfig.Password, cfg.DBConfig.Host, cfg.DBConfig.Port, cfg.DBConfig.DBName)
		ctx := context.Background()
		pool, err := postgresRepository.NewPool(ctx, dataSource)
		if err != nil {
			panic(err)
		}

		uow = postgresRepository.NewUnitOfWork(ctx, pool)
	default:
		panic(fmt.Sprintf("unknown storage chosen: %s", cfg.SrvConfig.Storage))
	}

	broker := subscription.NewBroker()

	commentsService := service.NewCommentsService(
		uow,
		broker,
	)

	graphqlHandler := graphql.NewHandler(
		commentsService,
		commentsService,
		broker,
	)

	port := ":" + cfg.SrvConfig.Port
	server := &http.Server{
		Addr:    port,
		Handler: graphqlHandler,
	}

	log.Printf("GraphQL server started on %s\n", port)

	if err := server.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
