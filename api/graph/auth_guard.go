package graph

import (
	"context"
	"strings"

	"setlist/auth"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// CodeUnauthenticated is the error extension code for an anonymous caller.
const CodeUnauthenticated = "UNAUTHENTICATED"

// RequireAuthenticated is a field middleware that denies anonymous callers on
// every root Query and Mutation field not marked @public, so a field added
// without the directive fails closed.
func RequireAuthenticated(ctx context.Context, next graphql.Resolver) (any, error) {
	fc := graphql.GetFieldContext(ctx)
	if fc == nil || (fc.Object != "Query" && fc.Object != "Mutation") {
		return next(ctx)
	}

	// Introspection fields such as __schema stay reachable for tooling.
	if strings.HasPrefix(fc.Field.Name, "__") {
		return next(ctx)
	}

	if fc.Field.Definition != nil && fc.Field.Definition.Directives.ForName("public") != nil {
		return next(ctx)
	}

	if _, ok := auth.UserIDFromContext(ctx); !ok {
		return nil, &gqlerror.Error{
			Message:    "Authentication required.",
			Extensions: map[string]any{"code": CodeUnauthenticated},
		}
	}
	return next(ctx)
}
