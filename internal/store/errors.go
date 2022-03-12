package store

import "errors"

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrNoRowsUpdated  = errors.New("no rows updated")
	ErrNoRowsDeleted  = errors.New("no records deleted")
)
