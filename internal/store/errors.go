package store

import "errors"

var (
	errRecordNotFound = errors.New("record not found")
	errNotUpdated = errors.New("not updated")
	errNoRowsAffected = errors.New("no rows affected")
)