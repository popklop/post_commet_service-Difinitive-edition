package repository

import (
	"context"

	"ozon-testovoe/internal/entity/comment"
	"ozon-testovoe/internal/entity/post"

	"github.com/google/uuid"
)

type PostRepos interface {
	Create(ctx context.Context, p *post.Post) error
	GetById(ctx context.Context, id uuid.UUID) (*post.Post, error)
	List(ctx context.Context, first int, after *string) ([]*post.Post, *string, bool, error)
	UpdateCommentsEnabled(ctx context.Context, id uuid.UUID, enabled bool) error
}

type CommentRepos interface {
	Create(ctx context.Context, c *comment.Comment) error
	GetRootComments(ctx context.Context, postID uuid.UUID, first int, after *string) ([]*comment.Comment, *string, bool, error)
	GetChildComments(ctx context.Context, parentID uuid.UUID, first int, after *string) ([]*comment.Comment, *string, bool, error)
	GetByID(ctx context.Context, id uuid.UUID) (*comment.Comment, error)
	GetByParentIDs(ctx context.Context, parentIDs []uuid.UUID, limit int) ([]*comment.Comment, error)
	GetRootsByPostIDs(ctx context.Context, postIDs []uuid.UUID, limit int) ([]*comment.Comment, error)
}
