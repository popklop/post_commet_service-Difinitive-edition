package comment

import (
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type Comment struct {
	createdAt time.Time
	parentID  *uuid.UUID
	text      string
	id        uuid.UUID
	author    uuid.UUID
	postID    uuid.UUID
}

func NewComment(text string, author uuid.UUID, parentID *uuid.UUID, postID uuid.UUID) (*Comment, error) {
	if text == "" {
		return nil, ErrEmptyCommentText
	}

	if utf8.RuneCountInString(text) > 2000 {
		return nil, ErrCommentTooLong
	}

	if author == uuid.Nil {
		return nil, ErrInvalidAuthor
	}

	if postID == uuid.Nil {
		return nil, ErrInvalidPost
	}

	now := time.Now()

	return &Comment{
		id:        uuid.New(),
		text:      text,
		author:    author,
		postID:    postID,
		parentID:  parentID,
		createdAt: now,
	}, nil
}

func (p *Comment) ID() uuid.UUID        { return p.id }
func (p *Comment) Text() string         { return p.text }
func (p *Comment) Author() uuid.UUID    { return p.author }
func (p *Comment) ParentID() *uuid.UUID { return p.parentID }
func (p *Comment) PostID() uuid.UUID    { return p.postID }
func (p *Comment) CreatedAt() time.Time { return p.createdAt }
func NewCommentFromDB(id uuid.UUID, text string, author, postID uuid.UUID, parentID *uuid.UUID, createdAt time.Time) *Comment {
	return &Comment{
		id:        id,
		text:      text,
		author:    author,
		postID:    postID,
		parentID:  parentID,
		createdAt: createdAt,
	}
}
