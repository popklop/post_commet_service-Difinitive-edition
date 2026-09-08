package memory

import (
	"context"
	"testing"

	"ozon-testovoe/internal/entity/post"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostRepo_Create(t *testing.T) {
	repo := NewPostRepository()
	ctx := context.Background()

	p, err := post.NewPost("Title", "Content", uuid.New(), true)
	require.NoError(t, err)

	err = repo.Create(ctx, p)
	assert.NoError(t, err)
	err = repo.Create(ctx, p)
	assert.ErrorIs(t, err, ErrAlreadyExistsPost)
}

func TestPostRepo_GetById(t *testing.T) {
	repo := NewPostRepository()
	ctx := context.Background()
	_, err := repo.GetById(ctx, uuid.New())
	assert.ErrorIs(t, err, ErrNotFoundPost)
	p, _ := post.NewPost("Title", "Content", uuid.New(), true)
	_ = repo.Create(ctx, p)
	found, err := repo.GetById(ctx, p.ID())
	assert.NoError(t, err)
	assert.Equal(t, p, found)
}

func TestPostRepo_List(t *testing.T) {
	repo := NewPostRepository()
	ctx := context.Background()
	posts, endCursor, hasNext, err := repo.List(ctx, 10, nil)
	assert.NoError(t, err)
	assert.Empty(t, posts)
	assert.Nil(t, endCursor)
	assert.False(t, hasNext)
	var ids []uuid.UUID
	for i := 0; i < 5; i++ {
		p, _ := post.NewPost("Title", "Content", uuid.New(), true)
		_ = repo.Create(ctx, p)
		ids = append(ids, p.ID())
	}
	posts, endCursor, hasNext, err = repo.List(ctx, 3, nil)
	assert.NoError(t, err)
	assert.Len(t, posts, 3)
	assert.NotNil(t, endCursor)
	assert.True(t, hasNext)
	assert.Equal(t, ids[4], posts[0].ID())
	assert.Equal(t, ids[3], posts[1].ID())
	assert.Equal(t, ids[2], posts[2].ID())
	posts2, endCursor2, hasNext2, err := repo.List(ctx, 3, endCursor)
	assert.NoError(t, err)
	assert.Len(t, posts2, 2)
	assert.NotNil(t, endCursor2)
	assert.False(t, hasNext2)
	assert.Equal(t, ids[1], posts2[0].ID())
	assert.Equal(t, ids[0], posts2[1].ID())
}

func TestPostRepo_UpdateCommentsEnabled(t *testing.T) {
	repo := NewPostRepository()
	ctx := context.Background()

	p, _ := post.NewPost("Title", "Content", uuid.New(), true)
	_ = repo.Create(ctx, p)

	err := repo.UpdateCommentsEnabled(ctx, p.ID(), false)
	assert.NoError(t, err)

	updated, _ := repo.GetById(ctx, p.ID())
	assert.False(t, updated.CommentsEnabled())
	err = repo.UpdateCommentsEnabled(ctx, uuid.New(), false)
	assert.ErrorIs(t, err, ErrNotFoundPost)
}
