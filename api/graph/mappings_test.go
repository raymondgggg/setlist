package graph

import (
	"setlist/db"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestToGraphQLUser(t *testing.T) {
	t.Run("db user fully populated", func(t *testing.T) {
		fn := "John"
		ln := "Doe"
		du := &db.User{
			ID:        uuid.MustParse("b03d292e-10f6-4e3f-9f51-383639acbbb0"),
			FirstName: &fn,
			LastName:  &ln,
			Email:     "jdoe@testmail.com",
		}
		expected := &User{
			ID:        uuid.MustParse("b03d292e-10f6-4e3f-9f51-383639acbbb0"),
			FirstName: &fn,
			LastName:  &ln,
			Email:     "jdoe@testmail.com",
		}

		u := ToGraphQLUser(du)
		assert.Equal(t, expected, u)
	})

	t.Run("first and last name are null", func(t *testing.T) {
		du := &db.User{
			ID:        uuid.MustParse("0a20840d-4036-47de-bd5a-8342e2d8f814"),
			FirstName: nil,
			LastName:  nil,
			Email:     "test@testmail.com",
		}
		expected := &User{
			ID:        uuid.MustParse("0a20840d-4036-47de-bd5a-8342e2d8f814"),
			FirstName: nil,
			LastName:  nil,
			Email:     "test@testmail.com",
		}

		u := ToGraphQLUser(du)
		assert.Equal(t, expected, u)
	})
}
