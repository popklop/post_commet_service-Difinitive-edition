package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"ozon-testovoe/internal/entity/post"
)

type mockPostRepo struct {
	createFunc         func(ctx context.Context, p *post.Post) error
	getByIDFunc        func(ctx context.Context, id uuid.UUID) (*post.Post, error)
	listFunc           func(ctx context.Context, first int, after *string) ([]*post.Post, *string, bool, error)
	updateCommentsFunc func(ctx context.Context, id uuid.UUID, enabled bool) error
}

func (m *mockPostRepo) Create(ctx context.Context, p *post.Post) error {
	return m.createFunc(ctx, p)
}
func (m *mockPostRepo) GetById(ctx context.Context, id uuid.UUID) (*post.Post, error) {
	return m.getByIDFunc(ctx, id)
}
func (m *mockPostRepo) List(ctx context.Context, first int, after *string) ([]*post.Post, *string, bool, error) {
	return m.listFunc(ctx, first, after)
}
func (m *mockPostRepo) UpdateCommentsEnabled(ctx context.Context, id uuid.UUID, enabled bool) error {
	return m.updateCommentsFunc(ctx, id, enabled)
}

func TestPostService_Create(t *testing.T) {
	ctx := context.Background()
	author := uuid.New()
	title := "Title"
	content := "Content"
	repo := &mockPostRepo{
		createFunc: func(ctx context.Context, p *post.Post) error {
			return nil
		},
	}
	svc := NewPostService(repo)
	p, err := svc.Create(ctx, title, content, author, true)
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, title, p.Title())
	assert.Equal(t, content, p.Content())
	assert.Equal(t, author, p.Author())
	assert.True(t, p.CommentsEnabled())
	repoErr := &mockPostRepo{
		createFunc: func(ctx context.Context, p *post.Post) error {
			return assert.AnError
		},
	}
	svcErr := NewPostService(repoErr)
	_, err = svcErr.Create(ctx, title, content, author, true)
	assert.Error(t, err)
}

func TestPostService_Get(t *testing.T) {
	ctx := context.Background()
	expectedPost, _ := post.NewPost("Title", "Content", uuid.New(), true)

	repo := &mockPostRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*post.Post, error) {
			if id == expectedPost.ID() {
				return expectedPost, nil
			}
			return nil, assert.AnError
		},
	}
	svc := NewPostService(repo)

	p, err := svc.Get(ctx, expectedPost.ID())
	assert.NoError(t, err)
	assert.Equal(t, expectedPost, p)
	_, err = svc.Get(ctx, uuid.New())
	assert.Error(t, err)
}

func TestPostService_List(t *testing.T) {
	ctx := context.Background()
	posts := []*post.Post{}
	for i := 0; i < 3; i++ {
		p, _ := post.NewPost("Title", "Content", uuid.New(), true)
		posts = append(posts, p)
	}
	endCursor := "cursor"
	hasNext := true

	repo := &mockPostRepo{
		listFunc: func(ctx context.Context, first int, after *string) ([]*post.Post, *string, bool, error) {
			return posts, &endCursor, hasNext, nil
		},
	}
	svc := NewPostService(repo)
	res, cursor, next, err := svc.List(ctx, 10, nil)
	assert.NoError(t, err)
	assert.Equal(t, posts, res)
	assert.Equal(t, &endCursor, cursor)
	assert.Equal(t, hasNext, next)
}

func TestPostService_UpdateCommentsEnabled(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	enabled := false
	updatedPost, _ := post.NewPost("Title", "Content", uuid.New(), true)
	updatedPost.SetCommentsEnabled(enabled)

	repo := &mockPostRepo{
		updateCommentsFunc: func(ctx context.Context, id uuid.UUID, enabled bool) error {
			return nil
		},
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*post.Post, error) {
			return updatedPost, nil
		},
	}
	svc := NewPostService(repo)
	p, err := svc.UpdateCommentsEnabled(ctx, id, enabled)
	assert.NoError(t, err)
	assert.Equal(t, updatedPost, p)
	repoErr := &mockPostRepo{
		updateCommentsFunc: func(ctx context.Context, id uuid.UUID, enabled bool) error {
			return assert.AnError
		},
	}
	svcErr := NewPostService(repoErr)
	_, err = svcErr.UpdateCommentsEnabled(ctx, id, enabled)
	assert.Error(t, err)
}
