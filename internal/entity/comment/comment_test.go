package comment

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewComment(t *testing.T) {
	author := uuid.New()
	postID := uuid.New()
	parentID := uuid.New()

	tests := []struct {
		expectedError error
		parentID      *uuid.UUID
		name          string
		text          string
		author        uuid.UUID
		postID        uuid.UUID
	}{
		{
			name:          "valid comment",
			text:          "Hello",
			author:        author,
			parentID:      nil,
			postID:        postID,
			expectedError: nil,
		},
		{
			name:          "valid comment with parent",
			text:          "Reply",
			author:        author,
			parentID:      &parentID,
			postID:        postID,
			expectedError: nil,
		},
		{
			name:          "empty text",
			text:          "",
			author:        author,
			parentID:      nil,
			postID:        postID,
			expectedError: ErrEmptyCommentText,
		},
		{
			name:          "too long text (>2000 runes)",
			text:          string(make([]rune, 2001)),
			author:        author,
			parentID:      nil,
			postID:        postID,
			expectedError: ErrCommentTooLong,
		},
		{
			name:          "nil author",
			text:          "Hello",
			author:        uuid.Nil,
			parentID:      nil,
			postID:        postID,
			expectedError: ErrInvalidAuthor,
		},
		{
			name:          "nil post",
			text:          "Hello",
			author:        author,
			parentID:      nil,
			postID:        uuid.Nil,
			expectedError: ErrInvalidPost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewComment(tt.text, tt.author, tt.parentID, tt.postID)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, c)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, c)
				assert.NotEqual(t, uuid.Nil, c.ID())
				assert.Equal(t, tt.text, c.Text())
				assert.Equal(t, tt.author, c.Author())
				assert.Equal(t, tt.postID, c.PostID())
				if tt.parentID == nil {
					assert.Nil(t, c.ParentID())
				} else {
					assert.Equal(t, *tt.parentID, *c.ParentID())
				}
				assert.WithinDuration(t, time.Now(), c.CreatedAt(), time.Second)
			}
		})
	}
}

func TestNewCommentFromDB(t *testing.T) {
	id := uuid.New()
	author := uuid.New()
	postID := uuid.New()
	parentID := uuid.New()
	now := time.Now()
	c := NewCommentFromDB(id, "Text", author, postID, &parentID, now)
	assert.Equal(t, id, c.ID())
	assert.Equal(t, "Text", c.Text())
	assert.Equal(t, author, c.Author())
	assert.Equal(t, postID, c.PostID())
	assert.Equal(t, parentID, *c.ParentID())
	assert.Equal(t, now, c.CreatedAt())
}
