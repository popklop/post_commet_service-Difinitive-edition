package post

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewPost(t *testing.T) {
	tests := []struct {
		expectedError   error
		name            string
		title           string
		content         string
		author          uuid.UUID
		commentsEnabled bool
	}{
		{
			name:            "valid post",
			title:           "Title",
			content:         "Content",
			author:          uuid.New(),
			commentsEnabled: true,
			expectedError:   nil,
		},
		{
			name:            "empty title",
			title:           "   ",
			content:         "Content",
			author:          uuid.New(),
			commentsEnabled: true,
			expectedError:   ErrEmptyPostTitle,
		},
		{
			name:            "empty content",
			title:           "Title",
			content:         "",
			author:          uuid.New(),
			commentsEnabled: true,
			expectedError:   ErrEmptyPostContent,
		},
		{
			name:            "nil author",
			title:           "Title",
			content:         "Content",
			author:          uuid.Nil,
			commentsEnabled: true,
			expectedError:   ErrInvalidAuthor,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPost(tt.title, tt.content, tt.author, tt.commentsEnabled)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, p)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, p)
				assert.NotEqual(t, uuid.Nil, p.ID())
				assert.Equal(t, tt.title, p.Title())
				assert.Equal(t, tt.content, p.Content())
				assert.Equal(t, tt.author, p.Author())
				assert.Equal(t, tt.commentsEnabled, p.CommentsEnabled())
				assert.WithinDuration(t, time.Now(), p.CreatedAt(), time.Second)
			}
		})
	}
}

func TestSetCommentsEnabled(t *testing.T) {
	p, _ := NewPost("Title", "Content", uuid.New(), true)
	p.SetCommentsEnabled(false)
	assert.False(t, p.CommentsEnabled())
}

func TestNewPostFromDB(t *testing.T) {
	id := uuid.New()
	author := uuid.New()
	now := time.Now()
	p := NewPostFromDB(id, "Title", "Content", author, true, now)
	assert.Equal(t, id, p.ID())
	assert.Equal(t, "Title", p.Title())
	assert.Equal(t, "Content", p.Content())
	assert.Equal(t, author, p.Author())
	assert.True(t, p.CommentsEnabled())
	assert.Equal(t, now, p.CreatedAt())
}
