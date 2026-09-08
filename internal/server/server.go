package server

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"net/http"
	"ozon-testovoe/graph/generated"
	"ozon-testovoe/graph/resolver"
	"ozon-testovoe/internal/dataloader"
	"ozon-testovoe/internal/service"
)

func NewServer(commentService service.CommentService, res *resolver.Resolver) http.Handler {
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: res}))

	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("GraphQL Playground", "/query"))
	mux.Handle("/query", dataLoaderMiddleware(commentService)(srv))
	return mux
}

func dataLoaderMiddleware(commentService service.CommentService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			loaders := dataloader.NewLoaders(commentService)
			ctx := dataloader.WithLoaders(r.Context(), loaders)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
