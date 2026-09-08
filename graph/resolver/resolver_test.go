package resolver

import (
	"context"
	"testing"

	"ozon-testovoe/graph/model"
	"ozon-testovoe/internal/entity/comment"
	"ozon-testovoe/internal/entity/post"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockPostService struct {
	createFunc         func(ctx context.Context, title, content string, author uuid.UUID, commentsEnabled bool) (*post.Post, error)
	getFunc            func(ctx context.Context, id uuid.UUID) (*post.Post, error)
	listFunc           func(ctx context.Context, first int, after *string) ([]*post.Post, *string, bool, error)
	updateCommentsFunc func(ctx context.Context, id uuid.UUID, enabled bool) (*post.Post, error)
}

func (m *mockPostService) Create(ctx context.Context, title, content string, author uuid.UUID, commentsEnabled bool) (*post.Post, error) {
	return m.createFunc(ctx, title, content, author, commentsEnabled)
}
func (m *mockPostService) Get(ctx context.Context, id uuid.UUID) (*post.Post, error) {
	return m.getFunc(ctx, id)
}
func (m *mockPostService) List(ctx context.Context, first int, after *string) ([]*post.Post, *string, bool, error) {
	return m.listFunc(ctx, first, after)
}
func (m *mockPostService) UpdateCommentsEnabled(ctx context.Context, id uuid.UUID, enabled bool) (*post.Post, error) {
	return m.updateCommentsFunc(ctx, id, enabled)
}

type mockCommentService struct {
	createFunc           func(ctx context.Context, text string, author, postID uuid.UUID, parentID *uuid.UUID) (*comment.Comment, error)
	getRootCommentsFunc  func(ctx context.Context, postID uuid.UUID, first int, after *string) ([]*comment.Comment, *string, bool, error)
	getChildCommentsFunc func(ctx context.Context, parentID uuid.UUID, first int, after *string) ([]*comment.Comment, *string, bool, error)
	subscribeFunc        func(ctx context.Context, postID uuid.UUID) chan *comment.Comment
	unsubscribeFunc      func(ctx context.Context, postID uuid.UUID, ch chan *comment.Comment)
	getChildrenBatchFunc func(ctx context.Context, parentIDs []uuid.UUID, limit int) (map[uuid.UUID][]*comment.Comment, error)
	getRootsBatchFunc    func(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID][]*comment.Comment, error)
}

func (m *mockCommentService) Create(ctx context.Context, text string, author, postID uuid.UUID, parentID *uuid.UUID) (*comment.Comment, error) {
	return m.createFunc(ctx, text, author, postID, parentID)
}
func (m *mockCommentService) GetRootComments(ctx context.Context, postID uuid.UUID, first int, after *string) ([]*comment.Comment, *string, bool, error) {
	return m.getRootCommentsFunc(ctx, postID, first, after)
}
func (m *mockCommentService) GetChildComments(ctx context.Context, parentID uuid.UUID, first int, after *string) ([]*comment.Comment, *string, bool, error) {
	return m.getChildCommentsFunc(ctx, parentID, first, after)
}
func (m *mockCommentService) Subscribe(ctx context.Context, postID uuid.UUID) chan *comment.Comment {
	return m.subscribeFunc(ctx, postID)
}
func (m *mockCommentService) Unsubscribe(ctx context.Context, postID uuid.UUID, ch chan *comment.Comment) {
	m.unsubscribeFunc(ctx, postID, ch)
}
func (m *mockCommentService) GetChildrenBatch(ctx context.Context, parentIDs []uuid.UUID, limit int) (map[uuid.UUID][]*comment.Comment, error) {
	if m.getChildrenBatchFunc != nil {
		return m.getChildrenBatchFunc(ctx, parentIDs, limit)
	}
	return make(map[uuid.UUID][]*comment.Comment), nil
}
func (m *mockCommentService) GetRootsBatch(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID][]*comment.Comment, error) {
	if m.getRootsBatchFunc != nil {
		return m.getRootsBatchFunc(ctx, postIDs)
	}
	return make(map[uuid.UUID][]*comment.Comment), nil
}

func TestResolver_Mutation_CreatePost(t *testing.T) {
	ctx := context.Background()
	input := model.CreatePostInput{
		Title:   "Title",
		Content: "Content",
		Author:  uuid.New().String(),
	}
	expectedPost, _ := post.NewPost(input.Title, input.Content, uuid.MustParse(input.Author), true)

	postSvc := &mockPostService{
		createFunc: func(ctx context.Context, title, content string, author uuid.UUID, commentsEnabled bool) (*post.Post, error) {
			return expectedPost, nil
		},
	}
	commentSvc := &mockCommentService{}
	resolver := &Resolver{PostService: postSvc, CommentService: commentSvc}
	mutRes := mutationResolver{Resolver: resolver}
	p, err := mutRes.CreatePost(ctx, input)
	assert.NoError(t, err)
	assert.Equal(t, expectedPost, p)
}

func TestResolver_Mutation_UpdatePostCommentsEnabled(t *testing.T) {
	ctx := context.Background()
	id := uuid.New().String()
	enabled := false
	expectedPost, _ := post.NewPost("Title", "Content", uuid.New(), true)
	expectedPost.SetCommentsEnabled(enabled)

	postSvc := &mockPostService{
		updateCommentsFunc: func(ctx context.Context, id uuid.UUID, enabled bool) (*post.Post, error) {
			return expectedPost, nil
		},
	}
	commentSvc := &mockCommentService{}
	resolver := &Resolver{PostService: postSvc, CommentService: commentSvc}
	mutRes := mutationResolver{Resolver: resolver}
	p, err := mutRes.UpdatePostCommentsEnabled(ctx, id, enabled)
	assert.NoError(t, err)
	assert.Equal(t, expectedPost, p)
}

func TestResolver_Mutation_CreateComment(t *testing.T) {
	ctx := context.Background()
	postID := uuid.New()
	author := uuid.New()
	input := model.CreateCommentInput{
		Text:   "Text",
		Author: author.String(),
		PostID: postID.String(),
	}
	expectedComment, _ := comment.NewComment(input.Text, author, nil, postID)

	commentSvc := &mockCommentService{
		createFunc: func(ctx context.Context, text string, author, postID uuid.UUID, parentID *uuid.UUID) (*comment.Comment, error) {
			return expectedComment, nil
		},
	}
	postSvc := &mockPostService{}
	resolver := &Resolver{PostService: postSvc, CommentService: commentSvc}
	mutRes := mutationResolver{Resolver: resolver}
	c, err := mutRes.CreateComment(ctx, input)
	assert.NoError(t, err)
	assert.Equal(t, expectedComment, c)
}

func TestResolver_Query_Posts(t *testing.T) {
	ctx := context.Background()
	first := 5
	after := "cursor"
	expectedPosts := []*post.Post{}
	endCursor := "end"
	hasNext := true

	postSvc := &mockPostService{
		listFunc: func(ctx context.Context, first int, after *string) ([]*post.Post, *string, bool, error) {
			return expectedPosts, &endCursor, hasNext, nil
		},
	}
	commentSvc := &mockCommentService{}
	resolver := &Resolver{PostService: postSvc, CommentService: commentSvc}
	qr := queryResolver{Resolver: resolver}
	conn, err := qr.Posts(ctx, &first, &after)
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	assert.Len(t, conn.Edges, len(expectedPosts))
	assert.Equal(t, endCursor, *conn.PageInfo.EndCursor)
	assert.Equal(t, hasNext, conn.PageInfo.HasNextPage)
}

func TestResolver_Query_Post(t *testing.T) {
	ctx := context.Background()
	id := uuid.New().String()
	expectedPost, _ := post.NewPost("Title", "Content", uuid.New(), true)

	postSvc := &mockPostService{
		getFunc: func(ctx context.Context, id uuid.UUID) (*post.Post, error) {
			return expectedPost, nil
		},
	}
	commentSvc := &mockCommentService{}
	resolver := &Resolver{PostService: postSvc, CommentService: commentSvc}
	qr := queryResolver{Resolver: resolver}
	p, err := qr.Post(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, expectedPost, p)
}
