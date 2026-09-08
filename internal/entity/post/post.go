package post

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Post struct {
	createdAt       time.Time
	title           string
	content         string
	id              uuid.UUID
	author          uuid.UUID
	commentsEnabled bool
}

func NewPost(title string, content string, author uuid.UUID, commentsEnabled bool) (*Post, error) {
	if len(strings.TrimSpace(title)) == 0 {
		return nil, ErrEmptyPostTitle
	}
	if len(strings.TrimSpace(content)) == 0 {
		return nil, ErrEmptyPostContent
	}
	if author == uuid.Nil {
		return nil, ErrInvalidAuthor
	}
	now := time.Now()
	return &Post{
		id:              uuid.New(),
		title:           title,
		content:         content,
		author:          author,
		commentsEnabled: commentsEnabled,
		createdAt:       now,
	}, nil
}

func (p *Post) ID() uuid.UUID         { return p.id }
func (p *Post) Title() string         { return p.title }
func (p *Post) Content() string       { return p.content }
func (p *Post) Author() uuid.UUID     { return p.author }
func (p *Post) CommentsEnabled() bool { return p.commentsEnabled }
func (p *Post) CreatedAt() time.Time  { return p.createdAt }

func (p *Post) SetCommentsEnabled(enabled bool) {
	p.commentsEnabled = enabled
}

func NewPostFromDB(id uuid.UUID, title, content string, author uuid.UUID, commentsEnabled bool, createdAt time.Time) *Post {
	return &Post{
		id:              id,
		title:           title,
		content:         content,
		author:          author,
		commentsEnabled: commentsEnabled,
		createdAt:       createdAt,
	}
}
