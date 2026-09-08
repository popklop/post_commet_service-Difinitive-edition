package memory

import (
	"context"
	"testing"

	"ozon-testovoe/internal/entity/comment"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommentRepo_Create(t *testing.T) {
	repo := NewCommentRepository()
	ctx := context.Background()

	c, err := comment.NewComment("Text", uuid.New(), nil, uuid.New())
	require.NoError(t, err)

	err = repo.Create(ctx, c)
	assert.NoError(t, err)
	err = repo.Create(ctx, c)
	assert.ErrorIs(t, err, ErrAlreadyExistsComm)
}

func TestCommentRepo_GetByID(t *testing.T) {
	repo := NewCommentRepository()
	ctx := context.Background()
	_, err := repo.GetByID(ctx, uuid.New())
	assert.ErrorIs(t, err, ErrNotFoundComm)

	c, _ := comment.NewComment("Text", uuid.New(), nil, uuid.New())
	_ = repo.Create(ctx, c)

	found, err := repo.GetByID(ctx, c.ID())
	assert.NoError(t, err)
	assert.Equal(t, c, found)
}

func TestCommentRepo_GetRootComments(t *testing.T) {
	repo := NewCommentRepository()
	ctx := context.Background()
	postID := uuid.New()
	parentID := uuid.New()
	var rootComments []*comment.Comment
	for i := 0; i < 5; i++ {
		c, _ := comment.NewComment("Root", uuid.New(), nil, postID)
		_ = repo.Create(ctx, c)
		rootComments = append(rootComments, c)
	}
	for i := 0; i < 3; i++ {
		c, _ := comment.NewComment("Child", uuid.New(), &parentID, postID)
		_ = repo.Create(ctx, c)
	}
	for i := 0; i < 2; i++ {
		c, _ := comment.NewComment("Other", uuid.New(), nil, uuid.New())
		_ = repo.Create(ctx, c)
	}

	comments, endCursor, hasNext, err := repo.GetRootComments(ctx, postID, 3, nil)
	assert.NoError(t, err)
	assert.Len(t, comments, 3)
	assert.NotNil(t, endCursor)
	assert.True(t, hasNext)
	assert.Equal(t, rootComments[4].ID(), comments[0].ID())
	assert.Equal(t, rootComments[3].ID(), comments[1].ID())
	assert.Equal(t, rootComments[2].ID(), comments[2].ID())

	comments2, endCursor2, hasNext2, err := repo.GetRootComments(ctx, postID, 3, endCursor)
	assert.NoError(t, err)
	assert.Len(t, comments2, 2)
	assert.NotNil(t, endCursor2)
	assert.False(t, hasNext2)
	assert.Equal(t, rootComments[1].ID(), comments2[0].ID())
	assert.Equal(t, rootComments[0].ID(), comments2[1].ID())
	_, _, _, err = repo.GetRootComments(ctx, postID, 3, stringPtr("invalid"))
	assert.Error(t, err)

	_, _, _, err = repo.GetRootComments(ctx, postID, 3, stringPtr(uuid.New().String()))
	assert.ErrorIs(t, err, ErrCursorNotFound)
}

func TestCommentRepo_GetChildComments(t *testing.T) {
	repo := NewCommentRepository()
	ctx := context.Background()
	parentID := uuid.New()
	postID := uuid.New()

	var childComments []*comment.Comment
	for i := 0; i < 5; i++ {
		c, _ := comment.NewComment("Child", uuid.New(), &parentID, postID)
		_ = repo.Create(ctx, c)
		childComments = append(childComments, c)
	}
	for i := 0; i < 3; i++ {
		c, _ := comment.NewComment("Other", uuid.New(), nil, postID)
		_ = repo.Create(ctx, c)
	}
	otherParent := uuid.New()
	for i := 0; i < 2; i++ {
		c, _ := comment.NewComment("OtherChild", uuid.New(), &otherParent, postID)
		_ = repo.Create(ctx, c)
	}

	comments, endCursor, hasNext, err := repo.GetChildComments(ctx, parentID, 3, nil)
	assert.NoError(t, err)
	assert.Len(t, comments, 3)
	assert.NotNil(t, endCursor)
	assert.True(t, hasNext)
	assert.Equal(t, childComments[4].ID(), comments[0].ID())
	assert.Equal(t, childComments[3].ID(), comments[1].ID())
	assert.Equal(t, childComments[2].ID(), comments[2].ID())

	comments2, endCursor2, hasNext2, err := repo.GetChildComments(ctx, parentID, 3, endCursor)
	assert.NoError(t, err)
	assert.Len(t, comments2, 2)
	assert.NotNil(t, endCursor2)
	assert.False(t, hasNext2)
	assert.Equal(t, childComments[1].ID(), comments2[0].ID())
	assert.Equal(t, childComments[0].ID(), comments2[1].ID())

	_, _, _, err = repo.GetChildComments(ctx, parentID, 3, stringPtr("invalid"))
	assert.Error(t, err)

	_, _, _, err = repo.GetChildComments(ctx, parentID, 3, stringPtr(uuid.New().String()))
	assert.ErrorIs(t, err, ErrCursorNotFound)
}

func stringPtr(s string) *string {
	return &s
}
