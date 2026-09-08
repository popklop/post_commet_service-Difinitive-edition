package service

import (
	"context"
	"testing"
	"time"

	"ozon-testovoe/internal/entity/comment"
	"ozon-testovoe/internal/entity/post"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockCommentRepo struct {
	createFunc            func(ctx context.Context, c *comment.Comment) error
	getRootCommentsFunc   func(ctx context.Context, postID uuid.UUID, first int, after *string) ([]*comment.Comment, *string, bool, error)
	getChildCommentsFunc  func(ctx context.Context, parentID uuid.UUID, first int, after *string) ([]*comment.Comment, *string, bool, error)
	getByIDFunc           func(ctx context.Context, id uuid.UUID) (*comment.Comment, error)
	getByParentIDsFunc    func(ctx context.Context, parentIDs []uuid.UUID, limit int) ([]*comment.Comment, error)
	getRootsByPostIDsFunc func(ctx context.Context, postIDs []uuid.UUID, limit int) ([]*comment.Comment, error)
}

func (m *mockCommentRepo) GetRootsByPostIDs(ctx context.Context, postIDs []uuid.UUID, limit int) ([]*comment.Comment, error) {
	if m.getRootsByPostIDsFunc != nil {
		return m.getRootsByPostIDsFunc(ctx, postIDs, limit)
	}
	return nil, nil
}

func (m *mockCommentRepo) Create(ctx context.Context, c *comment.Comment) error {
	return m.createFunc(ctx, c)
}
func (m *mockCommentRepo) GetRootComments(ctx context.Context, postID uuid.UUID, first int, after *string) ([]*comment.Comment, *string, bool, error) {
	return m.getRootCommentsFunc(ctx, postID, first, after)
}
func (m *mockCommentRepo) GetChildComments(ctx context.Context, parentID uuid.UUID, first int, after *string) ([]*comment.Comment, *string, bool, error) {
	return m.getChildCommentsFunc(ctx, parentID, first, after)
}
func (m *mockCommentRepo) GetByID(ctx context.Context, id uuid.UUID) (*comment.Comment, error) {
	return m.getByIDFunc(ctx, id)
}
func (m *mockCommentRepo) GetByParentIDs(ctx context.Context, parentIDs []uuid.UUID, limit int) ([]*comment.Comment, error) {
	if m.getByParentIDsFunc != nil {
		return m.getByParentIDsFunc(ctx, parentIDs, limit)
	}
	return nil, nil
}

type mockPostRepoForComment struct {
	getByIDFunc func(ctx context.Context, id uuid.UUID) (*post.Post, error)
}

func (m *mockPostRepoForComment) Create(ctx context.Context, p *post.Post) error { return nil }
func (m *mockPostRepoForComment) GetById(ctx context.Context, id uuid.UUID) (*post.Post, error) {
	return m.getByIDFunc(ctx, id)
}
func (m *mockPostRepoForComment) List(ctx context.Context, first int, after *string) ([]*post.Post, *string, bool, error) {
	return nil, nil, false, nil
}
func (m *mockPostRepoForComment) UpdateCommentsEnabled(ctx context.Context, id uuid.UUID, enabled bool) error {
	return nil
}

func TestCommentService_Create(t *testing.T) {
	ctx := context.Background()
	postID := uuid.New()
	author := uuid.New()
	text := "Hello"

	p, _ := post.NewPost("Title", "Content", uuid.New(), true)
	postObj := post.NewPostFromDB(postID, "Title", "Content", uuid.New(), true, p.CreatedAt())

	t.Run("success", func(t *testing.T) {
		postRepo := &mockPostRepoForComment{
			getByIDFunc: func(ctx context.Context, id uuid.UUID) (*post.Post, error) {
				return postObj, nil
			},
		}
		commentRepo := &mockCommentRepo{
			createFunc: func(ctx context.Context, c *comment.Comment) error {
				return nil
			},
			getByIDFunc: func(ctx context.Context, id uuid.UUID) (*comment.Comment, error) {
				return nil, assert.AnError
			},
		}
		svc := NewCommentService(postRepo, commentRepo)
		c, err := svc.Create(ctx, text, author, postID, nil)
		assert.NoError(t, err)
		assert.NotNil(t, c)
		assert.Equal(t, text, c.Text())
		assert.Equal(t, author, c.Author())
		assert.Equal(t, postID, c.PostID())
		assert.Nil(t, c.ParentID())
	})
	t.Run("parent comment from another post", func(t *testing.T) {
		anotherPostID := uuid.New()
		parentID := uuid.New()
		parentComment, _ := comment.NewComment("Parent", uuid.New(), nil, anotherPostID)

		postRepo := &mockPostRepoForComment{
			getByIDFunc: func(ctx context.Context, id uuid.UUID) (*post.Post, error) {
				return postObj, nil
			},
		}
		commentRepo := &mockCommentRepo{
			getByIDFunc: func(ctx context.Context, id uuid.UUID) (*comment.Comment, error) {
				return parentComment, nil
			},
		}
		svc := NewCommentService(postRepo, commentRepo)
		_, err := svc.Create(ctx, text, author, postID, &parentID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotPermittedComm)
	})
	t.Run("comments disabled", func(t *testing.T) {
		disabledPost := post.NewPostFromDB(postID, "Title", "Content", uuid.New(), false, time.Now())
		postRepo := &mockPostRepoForComment{
			getByIDFunc: func(ctx context.Context, id uuid.UUID) (*post.Post, error) {
				return disabledPost, nil
			},
		}
		commentRepo := &mockCommentRepo{
			createFunc: func(ctx context.Context, c *comment.Comment) error {
				return nil
			},
		}
		svc := NewCommentService(postRepo, commentRepo)
		_, err := svc.Create(ctx, text, author, postID, nil)
		assert.ErrorIs(t, err, ErrNotPermittedComm)
	})

	t.Run("post not found", func(t *testing.T) {
		postRepo := &mockPostRepoForComment{
			getByIDFunc: func(ctx context.Context, id uuid.UUID) (*post.Post, error) {
				return nil, assert.AnError
			},
		}
		commentRepo := &mockCommentRepo{}
		svc := NewCommentService(postRepo, commentRepo)
		_, err := svc.Create(ctx, text, author, postID, nil)
		assert.Error(t, err)
	})

	t.Run("parent comment not found", func(t *testing.T) {
		parentID := uuid.New()
		postRepo := &mockPostRepoForComment{
			getByIDFunc: func(ctx context.Context, id uuid.UUID) (*post.Post, error) {
				return postObj, nil
			},
		}
		commentRepo := &mockCommentRepo{
			getByIDFunc: func(ctx context.Context, id uuid.UUID) (*comment.Comment, error) {
				return nil, assert.AnError
			},
		}
		svc := NewCommentService(postRepo, commentRepo)
		_, err := svc.Create(ctx, text, author, postID, &parentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parent comment not found")
	})
}

func TestCommentService_GetRootComments(t *testing.T) {
	ctx := context.Background()
	postID := uuid.New()
	expectedComments := []*comment.Comment{}
	expectedCursor := "cursor"
	expectedHasNext := true

	commentRepo := &mockCommentRepo{
		getRootCommentsFunc: func(ctx context.Context, postID uuid.UUID, first int, after *string) ([]*comment.Comment, *string, bool, error) {
			return expectedComments, &expectedCursor, expectedHasNext, nil
		},
	}
	svc := NewCommentService(nil, commentRepo)
	comments, cursor, hasNext, err := svc.GetRootComments(ctx, postID, 10, nil)
	assert.NoError(t, err)
	assert.Equal(t, expectedComments, comments)
	assert.Equal(t, &expectedCursor, cursor)
	assert.Equal(t, expectedHasNext, hasNext)
}

func TestCommentService_GetChildComments(t *testing.T) {
	ctx := context.Background()
	parentID := uuid.New()
	expectedComments := []*comment.Comment{}
	expectedCursor := "cursor"
	expectedHasNext := true

	commentRepo := &mockCommentRepo{
		getChildCommentsFunc: func(ctx context.Context, parentID uuid.UUID, first int, after *string) ([]*comment.Comment, *string, bool, error) {
			return expectedComments, &expectedCursor, expectedHasNext, nil
		},
	}
	svc := NewCommentService(nil, commentRepo)
	comments, cursor, hasNext, err := svc.GetChildComments(ctx, parentID, 10, nil)
	assert.NoError(t, err)
	assert.Equal(t, expectedComments, comments)
	assert.Equal(t, &expectedCursor, cursor)
	assert.Equal(t, expectedHasNext, hasNext)
}

func TestCommentService_SubscribePublish(t *testing.T) {
	ctx := context.Background()
	postID := uuid.New()
	commentObj, _ := comment.NewComment(
		"Text",
		uuid.New(),
		nil,
		postID,
	)
	svc := NewCommentService(nil, nil).(*commentService)
	ch := svc.Subscribe(ctx, postID)
	defer svc.Unsubscribe(ctx, postID, ch)
	svc.publish(commentObj)
	select {
	case received := <-ch:
		assert.Equal(t, commentObj, received)
	default:
		t.Fatal("expected comment on channel")
	}
}

func TestCommentService_Unsubscribe(t *testing.T) {
	ctx := context.Background()
	postID := uuid.New()
	svc := NewCommentService(nil, nil).(*commentService)

	ch := svc.Subscribe(ctx, postID)
	svc.Unsubscribe(ctx, postID, ch)

	_, ok := <-ch
	assert.False(t, ok)

	svc.mu.RLock()
	defer svc.mu.RUnlock()
	_, exists := svc.subscribers[postID]
	assert.False(t, exists)
}
