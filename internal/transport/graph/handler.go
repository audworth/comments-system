package graph

import (
	"log/slog"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/audworth/comments-system/internal/transport/graph/dataloader"
	grapherror "github.com/audworth/comments-system/internal/transport/graph/error"
	"github.com/audworth/comments-system/internal/transport/graph/generated"
	"github.com/audworth/comments-system/internal/transport/graph/resolver"
	"github.com/vektah/gqlparser/v2/ast"
)

const (
	queryCacheSize   = 1000
	parserTokenLimit = 10000
)

type HandlerConfig struct {
	Local           bool
	ComplexityLimit int
}

func NewHandler(
	root *resolver.Resolver,
	users dataloader.UserReader,
	comments dataloader.CommentReader,
	logger *slog.Logger,
	config HandlerConfig,
) http.Handler {
	schemaConfig := generated.Config{Resolvers: root}
	configureQueryComplexity(&schemaConfig)

	graphql := handler.New(generated.NewExecutableSchema(schemaConfig))
	graphql.AddTransport(transport.Options{
		AllowedMethods: []string{http.MethodOptions, http.MethodGet, http.MethodPost},
	})
	graphql.AddTransport(transport.GET{UseGrapQLResponseJsonByDefault: true})
	graphql.AddTransport(transport.POST{UseGrapQLResponseJsonByDefault: true})
	graphql.SetQueryCache(lru.New[*ast.QueryDocument](queryCacheSize))
	graphql.SetParserTokenLimit(parserTokenLimit)
	graphql.Use(extension.FixedComplexityLimit(config.ComplexityLimit))
	graphql.SetErrorPresenter(grapherror.NewPresenter(logger).Present)
	graphql.SetRecoverFunc(grapherror.NewRecoverFunc(logger))

	mux := http.NewServeMux()
	handler := dataloader.Middleware(users, comments, graphql)

	mux.Handle("/query", handler)
	if config.Local {
		graphql.Use(extension.Introspection{})
		mux.Handle("GET /{$}", playground.Handler("playground", "/query"))
	}

	return mux
}
