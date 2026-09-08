package memory

import "errors"

var (
	ErrCursorNotFound    = errors.New("cursor not found")
	ErrAlreadyExistsComm = errors.New("comment already exists")
	ErrNotFoundComm      = errors.New("comment not found")
	ErrNotFoundPost      = errors.New("post doesnot exist")
	ErrAlreadyExistsPost = errors.New("post already exists")
)
