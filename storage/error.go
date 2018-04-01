package storage

import "fmt"

type DBError struct {
	Message string
}

func (e DBError) Error() string {
	return fmt.Sprintf("db error: %q", e.Message)
}

type NotFound struct {
	Entity, ID string
}

func (e NotFound) Error() string {
	return fmt.Sprintf("%q not found in %q", e.ID, e.Entity)
}

type UndefinedEntity struct {
	Entity string
}

func (e UndefinedEntity) Error() string {
	return fmt.Sprintf("entity %q is not defined", e.Entity)
}
