package comment

import "errors"

var (
	ErrInvalidPost      = errors.New("postID is invalid")
	ErrInvalidAuthor    = errors.New("authorID is invalid")
	ErrCommentTooLong   = errors.New("comment cant have more than 2000 symbols")
	ErrEmptyCommentText = errors.New("comment cant be empty")
)
