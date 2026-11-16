package graph

import (
	"context"
	"testing"

	"setlist/auth"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func fieldCtx(object, name string, directives ...string) context.Context {
	def := &ast.FieldDefinition{Name: name}
	for _, d := range directives {
		def.Directives = append(def.Directives, &ast.Directive{Name: d})
	}
	fc := &graphql.FieldContext{
		Object: object,
		Field:  graphql.CollectedField{Field: &ast.Field{Name: name, Definition: def}},
	}
	return graphql.WithFieldContext(context.Background(), fc)
}

func TestRequireAuthenticated(t *testing.T) {
	next := func(ctx context.Context) (any, error) { return "ok", nil }

	tests := []struct {
		name    string
		ctx     context.Context
		allowed bool
	}{
		{"anonymous denied on unmarked query", fieldCtx("Query", "session"), false},
		{"anonymous denied on unmarked mutation", fieldCtx("Mutation", "createReview"), false},
		{"anonymous allowed on public field", fieldCtx("Mutation", "login", "public"), true},
		{"anonymous allowed on introspection", fieldCtx("Query", "__schema"), true},
		{"anonymous allowed on nested field", fieldCtx("User", "email"), true},
		{"authenticated allowed on unmarked field", auth.WithUserID(fieldCtx("Query", "session"), uuid.New()), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RequireAuthenticated(tt.ctx, next)

			if tt.allowed {
				require.NoError(t, err)
				assert.Equal(t, "ok", got)
				return
			}

			var gerr *gqlerror.Error
			require.ErrorAs(t, err, &gerr)
			assert.Equal(t, CodeUnauthenticated, gerr.Extensions["code"])
			assert.Nil(t, got)
		})
	}
}
