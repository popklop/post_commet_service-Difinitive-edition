package service

import (
	"context"
	"fmt"
	"ozon-testovoe/internal/entity/comment"
	"ozon-testovoe/internal/repository"
	"sort"
	"sync"

	"github.com/google/uuid"
)

type commentService struct {
	postRepo    repository.PostRepos
	commentRepo repository.CommentRepos

	subscribers map[uuid.UUID][]chan *comment.Comment
	mu          sync.RWMutex
}

func NewCommentService(postRepo repository.PostRepos, commentRepo repository.CommentRepos) CommentService {
	return &commentService{
		postRepo:    postRepo,
		commentRepo: commentRepo,
		subscribers: make(map[uuid.UUID][]chan *comment.Comment),
	}
}

func (s *commentService) Subscribe(ctx context.Context, postID uuid.UUID) chan *comment.Comment {
	s.mu.Lock()
	defer s.mu.Unlock()
	ch := make(chan *comment.Comment, 10)
	s.subscribers[postID] = append(s.subscribers[postID], ch)
	return ch
}

func (s *commentService) Unsubscribe(ctx context.Context, postID uuid.UUID, ch chan *comment.Comment) {
	s.mu.Lock()
	defer s.mu.Unlock()
	subs, ok := s.subscribers[postID]
	if !ok {
		return
	}
	for i, c := range subs {
		if c == ch {
			s.subscribers[postID] = append(subs[:i], subs[i+1:]...)
			close(ch)
			break
		}
	}
	if len(s.subscribers[postID]) == 0 {
		delete(s.subscribers, postID)
	}
}

func (s *commentService) publish(c *comment.Comment) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, ch := range s.subscribers[c.PostID()] {
		select {
		case ch <- c:
		default:
		}
	}
}

func (s *commentService) Create(ctx context.Context, text string, author, postID uuid.UUID, parentID *uuid.UUID) (*comment.Comment, error) {
	p, err := s.postRepo.GetById(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("post not found: %w", err)
	}
	if !p.CommentsEnabled() {
		return nil, ErrNotPermittedComm
	}
	if parentID != nil {
		parent, parentErr := s.commentRepo.GetByID(ctx, *parentID)
		if parentErr != nil {
			return nil, fmt.Errorf("parent comment not found: %w", parentErr)
		}
		if parent.PostID() != postID {
			return nil, ErrNotPermittedComm
		}
	}
	c, err := comment.NewComment(text, author, parentID, postID)
	if err != nil {
		return nil, err
	}
	if err := s.commentRepo.Create(ctx, c); err != nil {
		return nil, err
	}

	go s.publish(c)

	return c, nil
}

func (s *commentService) GetRootComments(ctx context.Context, postID uuid.UUID, first int, after *string) ([]*comment.Comment, *string, bool, error) {
	return s.commentRepo.GetRootComments(ctx, postID, first, after)
}

func (s *commentService) GetChildComments(ctx context.Context, parentID uuid.UUID, first int, after *string) ([]*comment.Comment, *string, bool, error) {
	return s.commentRepo.GetChildComments(ctx, parentID, first, after)
}

func (s *commentService) GetChildrenBatch(ctx context.Context, parentIDs []uuid.UUID, limit int) (map[uuid.UUID][]*comment.Comment, error) {
	if len(parentIDs) == 0 || limit <= 0 {
		return make(map[uuid.UUID][]*comment.Comment), nil
	}
	comments, err := s.commentRepo.GetByParentIDs(ctx, parentIDs, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch children batch: %w", err)
	}
	result := make(map[uuid.UUID][]*comment.Comment)
	for _, c := range comments {
		if c.ParentID() == nil {
			continue
		}
		parentID := *c.ParentID()
		result[parentID] = append(result[parentID], c)
	}
	for _, children := range result {
		sort.Slice(children, func(i, j int) bool {
			return children[i].CreatedAt().After(children[j].CreatedAt())
		})
	}
	return result, nil
}

func (s *commentService) GetRootsBatch(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID][]*comment.Comment, error) {
	if len(postIDs) == 0 {
		return make(map[uuid.UUID][]*comment.Comment), nil
	}
	limit := 10
	comments, err := s.commentRepo.GetRootsByPostIDs(ctx, postIDs, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch root comments batch: %w", err)
	}
	result := make(map[uuid.UUID][]*comment.Comment)
	for _, c := range comments {
		result[c.PostID()] = append(result[c.PostID()], c)
	}
	return result, nil
}
