package memory

import (
	"context"
	"ozon-testovoe/internal/entity/post"
	"sort"
	"sync"

	"github.com/google/uuid"
)

type PostRepos struct {
	posts      map[uuid.UUID]*post.Post
	orderedIDs []uuid.UUID
	mu         sync.RWMutex
}

func NewPostRepository() *PostRepos {
	return &PostRepos{
		posts:      make(map[uuid.UUID]*post.Post),
		orderedIDs: []uuid.UUID{},
	}
}

func (postrepos *PostRepos) Create(ctx context.Context, p *post.Post) error {
	postrepos.mu.Lock()
	defer postrepos.mu.Unlock()
	if _, exists := postrepos.posts[p.ID()]; exists {
		return ErrAlreadyExistsPost
	}
	postrepos.posts[p.ID()] = p

	postrepos.orderedIDs = append(postrepos.orderedIDs, p.ID())
	sort.Slice(postrepos.orderedIDs, func(i, j int) bool {
		return postrepos.posts[postrepos.orderedIDs[i]].CreatedAt().After(postrepos.posts[postrepos.orderedIDs[j]].CreatedAt())
	})
	return nil

}
func (postreposp *PostRepos) GetById(ctx context.Context, id uuid.UUID) (*post.Post, error) {
	postreposp.mu.RLock()
	defer postreposp.mu.RUnlock()
	if _, exists := postreposp.posts[id]; !exists {
		return nil, ErrNotFoundPost
	}
	return postreposp.posts[id], nil
}

func (postrepos *PostRepos) List(ctx context.Context, first int, after *string) ([]*post.Post, *string, bool, error) {
	postrepos.mu.RLock()
	defer postrepos.mu.RUnlock()
	total := len(postrepos.orderedIDs)
	if total == 0 {
		return []*post.Post{}, nil, false, nil
	}
	var afterval uuid.UUID
	hasCursor := false
	if after != nil {
		var err error
		afterval, err = uuid.Parse(*after)
		if err != nil {
			return nil, nil, false, err
		}
		hasCursor = true
	}
	posts := make([]*post.Post, 0, first)
	tmp := !hasCursor
	k := 0
	added := 0

	for i := 0; i < total; i++ {
		id := postrepos.orderedIDs[i]
		if hasCursor && id == afterval {
			tmp = true
			continue
		}
		if !tmp {
			continue
		}
		k++
		if added < first {
			posts = append(posts, postrepos.posts[id])
			added++
		}
	}
	if hasCursor && !tmp {
		return nil, nil, false, ErrCursorNotFound
	}

	haveWeMoreAfter := k > first
	var endCursor *string
	if len(posts) > 0 {
		cursor := posts[len(posts)-1].ID().String()
		endCursor = &cursor
	}
	return posts, endCursor, haveWeMoreAfter, nil
}
func (postrepos *PostRepos) UpdateCommentsEnabled(ctx context.Context, id uuid.UUID, enabled bool) error {
	postrepos.mu.Lock()
	defer postrepos.mu.Unlock()
	p, exists := postrepos.posts[id]
	if !exists {
		return ErrNotFoundPost
	}
	p.SetCommentsEnabled(enabled)
	return nil
}
