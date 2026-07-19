package graph

import (
	"log/slog"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	grapherror "github.com/audworth/comments-system/internal/transport/graph/error"
	"github.com/audworth/comments-system/internal/transport/graph/generated"
	"github.com/audworth/comments-system/internal/transport/graph/resolver"
	"github.com/vektah/gqlparser/v2/ast"
)

const (
	queryCacheSize     = 1000
	queryComplexity    = 200
	parserTokenLimit   = 10_000
	maxRequestBodySize = 1 << 20
)

func NewHandler(root *resolver.Resolver, logger *slog.Logger, local bool) http.Handler {
	graphql := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: root}))
	graphql.AddTransport(transport.Options{
		AllowedMethods: []string{http.MethodOptions, http.MethodGet, http.MethodPost},
	})
	graphql.AddTransport(transport.GET{UseGrapQLResponseJsonByDefault: true})
	graphql.AddTransport(transport.POST{UseGrapQLResponseJsonByDefault: true})
	graphql.SetQueryCache(lru.New[*ast.QueryDocument](queryCacheSize))
	graphql.SetParserTokenLimit(parserTokenLimit)
	graphql.Use(extension.FixedComplexityLimit(queryComplexity))
	graphql.SetErrorPresenter(grapherror.NewPresenter(logger).Present)
	graphql.SetRecoverFunc(grapherror.NewRecoverFunc(logger))

	mux := http.NewServeMux()
	mux.Handle("/query", http.MaxBytesHandler(graphql, maxRequestBodySize))
	if local {
		graphql.Use(extension.Introspection{})
		mux.Handle("GET /{$}", playground.Handler("playground", "/query"))
	}

	return mux
}
