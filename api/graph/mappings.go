package graph

import (
	"setlist/db"
)

func ToGraphQLUser(du *db.User) *User {
	return &User{
		ID:        du.ID,
		FirstName: du.FirstName,
		LastName:  du.LastName,
		Email:     du.Email,
	}
}
