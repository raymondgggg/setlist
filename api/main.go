package main

import (
	"setlist/config"
	"setlist/db"
	"setlist/graph"
	resolvers "setlist/graph/resolvers"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/vektah/gqlparser/v2/ast"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Defining the Graphql handler
func graphqlHandler(ur *db.UserRepository) gin.HandlerFunc {
	r := resolvers.NewResolver(ur)

	// NewExecutableSchema and Config are in the generated.go file
	// Resolver is in graph/resolvers/resolver.go
	h := handler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers: r,
	}))

	// Server setup:
	h.AddTransport(transport.Options{})
	h.AddTransport(transport.GET{})
	h.AddTransport(transport.POST{})

	h.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	h.Use(extension.Introspection{})
	h.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// Defining the Playground handler
func playgroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL", "/query")

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func main() {
	logger := zap.Must(zap.NewDevelopment())
	defer logger.Sync() // flushes buffer, if any

	c, err := config.LoadEnv()
	if err != nil {
		panic("cannot load env config")
	}

	gormDB, err := gorm.Open(postgres.Open(c.DBUrl))
	if err != nil {
		panic("error opening gorm connection")
	}

	// Create repositories
	ur := db.NewUserRepository(gormDB)

	// Setting up Gin
	r := gin.Default()
	r.Use(gin.Recovery())
	r.POST("/query", graphqlHandler(ur))
	r.GET("/", playgroundHandler())
	r.Run()
}
