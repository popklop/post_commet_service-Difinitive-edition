package service

import (
	"context"
	"ozon-testovoe/internal/entity/comment"
	"ozon-testovoe/internal/entity/post"

	"github.com/google/uuid"
)

type PostService interface {
	Create(ctx context.Context, title, content string, author uuid.UUID, commentsEnabled bool) (*post.Post, error)
	Get(ctx context.Context, id uuid.UUID) (*post.Post, error)
	List(ctx context.Context, first int, after *string) ([]*post.Post, *string, bool, error)
	UpdateCommentsEnabled(ctx context.Context, id uuid.UUID, enabled bool) (*post.Post, error)
}

type CommentService interface {
	Create(ctx context.Context, text string, author, postID uuid.UUID, parentID *uuid.UUID) (*comment.Comment, error)
	GetRootComments(ctx context.Context, postID uuid.UUID, first int, after *string) ([]*comment.Comment, *string, bool, error)
	GetChildComments(ctx context.Context, parentID uuid.UUID, first int, after *string) ([]*comment.Comment, *string, bool, error)
	Subscribe(ctx context.Context, postID uuid.UUID) chan *comment.Comment
	Unsubscribe(ctx context.Context, postID uuid.UUID, ch chan *comment.Comment)
	GetChildrenBatch(ctx context.Context, parentIDs []uuid.UUID, limit int) (map[uuid.UUID][]*comment.Comment, error)
	GetRootsBatch(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID][]*comment.Comment, error)
}
