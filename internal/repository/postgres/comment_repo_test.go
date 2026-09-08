//go:build integration
// +build integration

package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"ozon-testovoe/internal/entity/comment"
	"ozon-testovoe/internal/entity/post"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func stringPtr(s string) *string {
	return &s
}

func TestCommentRepos_Create(t *testing.T) {
	ctx := context.Background()
	_, err := testDB.ExecContext(ctx, "DELETE FROM comments; DELETE FROM posts;")
	require.NoError(t, err)

	postRepo := NewPostRepos(testDB)
	commentRepo := NewCommentRepos(testDB)

	p, err := post.NewPost("Title", "Content", uuid.New(), true)
	require.NoError(t, err)
	err = postRepo.Create(ctx, p)
	require.NoError(t, err)

	c, err := comment.NewComment("Text", uuid.New(), nil, p.ID())
	require.NoError(t, err)

	err = commentRepo.Create(ctx, c)
	assert.NoError(t, err)

	var count int
	err = testDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM comments WHERE id = $1", c.ID()).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestCommentRepos_GetByID(t *testing.T) {
	ctx := context.Background()
	_, err := testDB.ExecContext(ctx, "DELETE FROM comments; DELETE FROM posts;")
	require.NoError(t, err)

	postRepo := NewPostRepos(testDB)
	commentRepo := NewCommentRepos(testDB)

	p, err := post.NewPost("Title", "Content", uuid.New(), true)
	require.NoError(t, err)
	err = postRepo.Create(ctx, p)
	require.NoError(t, err)

	c, err := comment.NewComment("Text", uuid.New(), nil, p.ID())
	require.NoError(t, err)
	err = commentRepo.Create(ctx, c)
	require.NoError(t, err)

	found, err := commentRepo.GetByID(ctx, c.ID())
	assert.NoError(t, err)
	assert.Equal(t, c.ID(), found.ID())
	assert.Equal(t, c.Text(), found.Text())
	assert.Equal(t, c.Author(), found.Author())
	assert.Equal(t, c.PostID(), found.PostID())
	assert.Nil(t, found.ParentID())
	assert.WithinDuration(t, c.CreatedAt(), found.CreatedAt(), time.Second)

	_, err = commentRepo.GetByID(ctx, uuid.New())
	assert.ErrorIs(t, err, ErrCommNotFound)
}

func TestCommentRepos_GetRootComments(t *testing.T) {
	ctx := context.Background()
	_, err := testDB.ExecContext(ctx, "DELETE FROM comments; DELETE FROM posts;")
	require.NoError(t, err)

	postRepo := NewPostRepos(testDB)
	commentRepo := NewCommentRepos(testDB)

	postID := uuid.New()
	p := post.NewPostFromDB(postID, "Title", "Content", uuid.New(), true, time.Now())
	err = postRepo.Create(ctx, p)
	require.NoError(t, err)
	var rootIDs []uuid.UUID
	for i := 0; i < 5; i++ {
		c, _ := comment.NewComment(fmt.Sprintf("Root %d", i), uuid.New(), nil, postID)
		_ = commentRepo.Create(ctx, c)
		rootIDs = append(rootIDs, c.ID())
	}
	parentID := rootIDs[0]
	for i := 0; i < 3; i++ {
		c, _ := comment.NewComment(fmt.Sprintf("Child %d", i), uuid.New(), &parentID, postID)
		_ = commentRepo.Create(ctx, c)
	}
	comments, endCursor, hasNext, err := commentRepo.GetRootComments(ctx, postID, 3, nil)
	assert.NoError(t, err)
	assert.Len(t, comments, 3)
	assert.NotNil(t, endCursor)
	assert.True(t, hasNext)
	for i := 1; i < len(comments); i++ {
		assert.True(t, comments[i-1].CreatedAt().After(comments[i].CreatedAt()) || comments[i-1].CreatedAt().Equal(comments[i].CreatedAt()))
	}
	comments2, endCursor2, hasNext2, err := commentRepo.GetRootComments(ctx, postID, 3, endCursor)
	assert.NoError(t, err)
	assert.Len(t, comments2, 2)
	assert.NotNil(t, endCursor2)
	assert.False(t, hasNext2)
	_, _, _, err = commentRepo.GetRootComments(ctx, postID, 3, stringPtr("invalid"))
	assert.ErrorIs(t, err, ErrInvalidCursor)
}

func TestCommentRepos_GetChildComments(t *testing.T) {
	ctx := context.Background()
	_, err := testDB.ExecContext(ctx, "DELETE FROM comments; DELETE FROM posts;")
	require.NoError(t, err)

	postRepo := NewPostRepos(testDB)
	commentRepo := NewCommentRepos(testDB)

	postID := uuid.New()
	p := post.NewPostFromDB(postID, "Title", "Content", uuid.New(), true, time.Now())
	err = postRepo.Create(ctx, p)
	require.NoError(t, err)
	parentComment, err := comment.NewComment("Parent", uuid.New(), nil, postID)
	require.NoError(t, err)
	err = commentRepo.Create(ctx, parentComment)
	require.NoError(t, err)
	parentID := parentComment.ID()
	var childIDs []uuid.UUID
	for i := 0; i < 5; i++ {
		c, _ := comment.NewComment(fmt.Sprintf("Child %d", i), uuid.New(), &parentID, postID)
		_ = commentRepo.Create(ctx, c)
		childIDs = append(childIDs, c.ID())
	}

	comments, endCursor, hasNext, err := commentRepo.GetChildComments(ctx, parentID, 3, nil)
	assert.NoError(t, err)
	assert.Len(t, comments, 3)
	assert.NotNil(t, endCursor)
	assert.True(t, hasNext)
	for i := 1; i < len(comments); i++ {
		assert.True(t, comments[i-1].CreatedAt().After(comments[i].CreatedAt()) || comments[i-1].CreatedAt().Equal(comments[i].CreatedAt()))
	}

	comments2, endCursor2, hasNext2, err := commentRepo.GetChildComments(ctx, parentID, 3, endCursor)
	assert.NoError(t, err)
	assert.Len(t, comments2, 2)
	assert.NotNil(t, endCursor2)
	assert.False(t, hasNext2)

	_, _, _, err = commentRepo.GetChildComments(ctx, parentID, 3, stringPtr("invalid"))
	assert.ErrorIs(t, err, ErrInvalidCursor)
}
