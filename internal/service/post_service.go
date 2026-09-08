package service

import (
	"context"
	"fmt"
	"ozon-testovoe/internal/entity/post"
	"ozon-testovoe/internal/repository"

	"github.com/google/uuid"
)

type postService struct {
	repo repository.PostRepos
}

func NewPostService(repo repository.PostRepos) PostService {
	return &postService{repo: repo}
}

func (s *postService) Create(ctx context.Context, title, content string, author uuid.UUID, commentsEnabled bool) (*post.Post, error) {
	p, err := post.NewPost(title, content, author, commentsEnabled)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}
	return p, nil
}

func (s *postService) Get(ctx context.Context, id uuid.UUID) (*post.Post, error) {
	p, err := s.repo.GetById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}
	return p, nil
}

func (s *postService) List(ctx context.Context, first int, after *string) ([]*post.Post, *string, bool, error) {
	posts, endCursor, hasNext, err := s.repo.List(ctx, first, after)
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to list posts: %w", err)
	}
	return posts, endCursor, hasNext, nil
}

func (s *postService) UpdateCommentsEnabled(ctx context.Context, id uuid.UUID, enabled bool) (*post.Post, error) {
	if err := s.repo.UpdateCommentsEnabled(ctx, id, enabled); err != nil {
		return nil, fmt.Errorf("failed to update comments enabled: %w", err)
	}
	p, err := s.repo.GetById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated post: %w", err)
	}
	return p, nil
}
