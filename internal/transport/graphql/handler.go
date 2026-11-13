package graphql

import (
	"net/http"

	"comments-service/internal/transport/graphql/generated"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
)

func NewHandler(posts PostService, comments CommentService, subscriber CommentSubscriber) http.Handler {
	resolver := NewResolver(posts, comments, subscriber)

	schema := generated.NewExecutableSchema(
		generated.Config{
			Resolvers: resolver,
		},
	)

	srv := handler.New(schema)

	srv.Use(extension.FixedComplexityLimit(100))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.AddTransport(transport.Websocket{})

	mux := http.NewServeMux()

	mux.Handle("/graphql", srv)

	mux.Handle("/", playground.Handler("GraphQL Playground", "/graphql"))

	return mux
}
