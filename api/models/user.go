package models

import "github.com/google/uuid"

type User struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Email     string
}
