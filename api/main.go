package main

import (
	"context"
	"setlist/auth"
	"setlist/config"
	"setlist/db"
	"setlist/email"
	"setlist/graph"
	resolvers "setlist/graph/resolvers"
	"setlist/middleware"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/resend/resend-go/v4"
	"github.com/vektah/gqlparser/v2/ast"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Defining the Graphql handler
func graphqlHandler(ur *db.UserRepository, as *auth.Service, l *zap.Logger, e config.Environment) gin.HandlerFunc {
	r := resolvers.NewResolver(ur, as, l, e)

	// NewExecutableSchema and Config are in the generated.go file
	// Resolver is in graph/resolvers/resolver.go
	h := handler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers: r,
	}))

	// Server setup:
	h.AddTransport(transport.Options{})
	h.AddTransport(transport.GET{})
	h.AddTransport(transport.POST{})

	h.AroundFields(graph.RequireAuthenticated)

	h.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	h.Use(extension.Introspection{})
	h.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		ctx = graph.WithResponseWriter(ctx, c.Writer)
		ctx = graph.WithRequestReader(ctx, c.Request)

		r := c.Request.WithContext(ctx)

		h.ServeHTTP(c.Writer, r)
	}
}

// Defining the Playground handler
func playgroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL", "/graphql")

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
	sr := db.NewSessionRepository(gormDB)
	vr := db.NewEmailVerificationRepository(gormDB)

	// Create email sender
	var ec interface {
		SendUserVerification(ctx context.Context, to, token string) error
	}
	if c.EmailLogOnly {
		logger.Warn("EMAIL_LOG_ONLY is enabled: emails will be logged, not sent")
		ec = email.NewLogSender(logger, c.FrontendBaseURL)
	} else {
		ec = email.NewClient(resend.NewClient(c.ResendAPIKey), logger, c.EmailFrom, c.FrontendBaseURL)
	}

	// Create auth service
	as := auth.NewService(ur, sr, vr, ec, logger, c.JWTSecret)

	// Setting up Gin
	r := gin.Default()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS(c.FrontendBaseURL))
	r.POST("/graphql", middleware.Authenticate([]byte(c.JWTSecret)), graphqlHandler(ur, as, logger, c.Env))
	r.GET("/", playgroundHandler())
	r.Run()
}
