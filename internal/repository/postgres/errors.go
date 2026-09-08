package postgres

import "errors"

var (
	ErrCommNotFound    = errors.New("comment not found")
	ErrNegativeFirst   = errors.New("first must be positive")
	ErrInvalidCursor   = errors.New("invalid cursor")
	ErrCursorNotFound  = errors.New("cursor not found")
	ErrPostNotOnUpdate = errors.New("post with that id not found, changes not done")
	ErrPostNotFound    = errors.New("post with thay id doesnt exist")
)
