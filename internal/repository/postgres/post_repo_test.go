//go:build integration
// +build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"ozon-testovoe/internal/entity/post"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostRepos_Create(t *testing.T) {
	ctx := context.Background()
	_, err := testDB.ExecContext(ctx, "DELETE FROM comments; DELETE FROM posts;")
	require.NoError(t, err)

	postRepo := NewPostRepos(testDB)

	p, err := post.NewPost("Title", "Content", uuid.New(), true)
	require.NoError(t, err)

	err = postRepo.Create(ctx, p)
	assert.NoError(t, err)

	var count int
	err = testDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM posts WHERE id = $1", p.ID()).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestPostRepos_GetById(t *testing.T) {
	ctx := context.Background()
	_, err := testDB.ExecContext(ctx, "DELETE FROM comments; DELETE FROM posts;")
	require.NoError(t, err)

	postRepo := NewPostRepos(testDB)

	p, err := post.NewPost("Title", "Content", uuid.New(), true)
	require.NoError(t, err)
	err = postRepo.Create(ctx, p)
	require.NoError(t, err)

	found, err := postRepo.GetById(ctx, p.ID())
	assert.NoError(t, err)
	assert.Equal(t, p.ID(), found.ID())
	assert.Equal(t, p.Title(), found.Title())
	assert.Equal(t, p.Content(), found.Content())
	assert.Equal(t, p.Author(), found.Author())
	assert.Equal(t, p.CommentsEnabled(), found.CommentsEnabled())
	assert.WithinDuration(t, p.CreatedAt(), found.CreatedAt(), time.Second)

	_, err = postRepo.GetById(ctx, uuid.New())
	assert.Error(t, err)
}

func TestPostRepos_List(t *testing.T) {
	ctx := context.Background()
	_, err := testDB.ExecContext(ctx, "DELETE FROM comments; DELETE FROM posts;")
	require.NoError(t, err)
	postRepo := NewPostRepos(testDB)
	var ids []uuid.UUID
	for i := 0; i < 5; i++ {
		p, _ := post.NewPost("Title", "Content", uuid.New(), true)
		_ = postRepo.Create(ctx, p)
		ids = append(ids, p.ID())
	}
	posts, endCursor, hasNext, err := postRepo.List(ctx, 3, nil)
	assert.NoError(t, err)
	assert.Len(t, posts, 3)
	assert.NotNil(t, endCursor)
	assert.True(t, hasNext)
	for i := 1; i < len(posts); i++ {
		assert.True(t, posts[i-1].CreatedAt().After(posts[i].CreatedAt()) || posts[i-1].CreatedAt().Equal(posts[i].CreatedAt()))
	}
	posts2, endCursor2, hasNext2, err := postRepo.List(ctx, 3, endCursor)
	assert.NoError(t, err)
	assert.Len(t, posts2, 2)
	assert.NotNil(t, endCursor2)
	assert.False(t, hasNext2)
}

func TestPostRepos_UpdateCommentsEnabled(t *testing.T) {
	ctx := context.Background()
	_, err := testDB.ExecContext(ctx, "DELETE FROM comments; DELETE FROM posts;")
	require.NoError(t, err)

	postRepo := NewPostRepos(testDB)

	p, err := post.NewPost("Title", "Content", uuid.New(), true)
	require.NoError(t, err)
	err = postRepo.Create(ctx, p)
	require.NoError(t, err)

	err = postRepo.UpdateCommentsEnabled(ctx, p.ID(), false)
	assert.NoError(t, err)

	updated, err := postRepo.GetById(ctx, p.ID())
	assert.NoError(t, err)
	assert.False(t, updated.CommentsEnabled())

	err = postRepo.UpdateCommentsEnabled(ctx, uuid.New(), false)
	assert.ErrorIs(t, err, ErrPostNotOnUpdate)
}
