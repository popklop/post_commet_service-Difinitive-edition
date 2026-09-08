package post

import "errors"

var (
	ErrInvalidAuthor    = errors.New("authorID is invalid")
	ErrEmptyPostContent = errors.New("post content cant be empty")
	ErrEmptyPostTitle   = errors.New("post title cant be empty")
)
