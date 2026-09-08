package service

import "errors"

var (
	ErrNotPermittedComm = errors.New("comments are disabled for this post")
)
