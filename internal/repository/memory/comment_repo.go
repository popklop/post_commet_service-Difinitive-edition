package memory

import (
	"context"
	"ozon-testovoe/internal/entity/comment"
	"sort"
	"sync"

	"github.com/google/uuid"
)

type CommentRepos struct {
	comments        map[uuid.UUID]*comment.Comment
	orderedComments []uuid.UUID
	mu              sync.RWMutex
}

func NewCommentRepository() *CommentRepos {
	return &CommentRepos{
		comments:        make(map[uuid.UUID]*comment.Comment),
		orderedComments: []uuid.UUID{},
	}
}

func (comrepos *CommentRepos) Create(ctx context.Context, c *comment.Comment) error {
	comrepos.mu.Lock()
	defer comrepos.mu.Unlock()
	if _, exists := comrepos.comments[c.ID()]; !exists {
		comrepos.comments[c.ID()] = c
		comrepos.orderedComments = append(comrepos.orderedComments, c.ID())
		return nil
	}
	return ErrAlreadyExistsComm
}

func (comrepos *CommentRepos) GetRootComments(ctx context.Context, postID uuid.UUID, first int, after *string) ([]*comment.Comment, *string, bool, error) {
	comrepos.mu.RLock()
	defer comrepos.mu.RUnlock()

	all := make([]*comment.Comment, 0)
	for _, c := range comrepos.comments {
		if c.PostID() == postID && c.ParentID() == nil {
			all = append(all, c)
		}
	}
	if len(all) == 0 {
		return []*comment.Comment{}, nil, false, nil
	}
	sortComments(all)

	startIdx := 0
	if after != nil {
		afterVal, err := uuid.Parse(*after)
		if err != nil {
			return nil, nil, false, err
		}
		found := false
		for i, c := range all {
			if c.ID() == afterVal {
				startIdx = i + 1
				found = true
				break
			}
		}
		if !found {
			return nil, nil, false, ErrCursorNotFound
		}
	}
	if startIdx >= len(all) {
		return []*comment.Comment{}, nil, false, nil
	}
	endIdx := startIdx + first
	if endIdx > len(all) {
		endIdx = len(all)
	}
	result := all[startIdx:endIdx]
	haveWeMoreAfter := endIdx < len(all)
	var lastused *string
	if len(result) > 0 {
		cursor := result[len(result)-1].ID().String()
		lastused = &cursor
	}

	return result, lastused, haveWeMoreAfter, nil
}

func (comrepos *CommentRepos) GetChildComments(ctx context.Context, parentID uuid.UUID, first int, after *string) ([]*comment.Comment, *string, bool, error) {
	comrepos.mu.RLock()
	defer comrepos.mu.RUnlock()
	all := make([]*comment.Comment, 0)
	for _, c := range comrepos.comments {
		if c.ParentID() != nil && *c.ParentID() == parentID {
			all = append(all, c)
		}
	}

	if len(all) == 0 {
		return []*comment.Comment{}, nil, false, nil
	}
	sortComments(all)
	startIdx := 0

	if after != nil {
		afterVal, err := uuid.Parse(*after)
		if err != nil {
			return nil, nil, false, err
		}
		found := false
		for i, c := range all {
			if c.ID() == afterVal {
				startIdx = i + 1
				found = true
				break
			}
		}
		if !found {
			return nil, nil, false, ErrCursorNotFound
		}
	}

	if startIdx >= len(all) {
		return []*comment.Comment{}, nil, false, nil
	}
	endIdx := startIdx + first
	if endIdx > len(all) {
		endIdx = len(all)
	}
	result := all[startIdx:endIdx]
	haveWeMoreAfter := endIdx < len(all)
	var lastfound *string
	if len(result) > 0 {
		cursor := result[len(result)-1].ID().String()
		lastfound = &cursor
	}

	return result, lastfound, haveWeMoreAfter, nil
}

func (comrepos *CommentRepos) GetByID(ctx context.Context, id uuid.UUID) (*comment.Comment, error) {
	comrepos.mu.RLock()
	defer comrepos.mu.RUnlock()
	c, exists := comrepos.comments[id]
	if !exists {
		return nil, ErrNotFoundComm
	}
	return c, nil
}

func (comrepos *CommentRepos) GetByParentIDs(ctx context.Context, parentIDs []uuid.UUID, limit int) ([]*comment.Comment, error) {
	comrepos.mu.RLock()
	defer comrepos.mu.RUnlock()

	if len(parentIDs) == 0 || limit <= 0 {
		return []*comment.Comment{}, nil
	}
	parentSet := make(map[uuid.UUID]struct{}, len(parentIDs))
	for _, id := range parentIDs {
		parentSet[id] = struct{}{}
	}
	byParent := make(map[uuid.UUID][]*comment.Comment, len(parentIDs))

	for _, c := range comrepos.comments {
		if c.ParentID() == nil {
			continue
		}
		parentID := *c.ParentID()

		if _, ok := parentSet[parentID]; !ok {
			continue
		}
		byParent[parentID] = append(byParent[parentID], c)
	}
	result := make([]*comment.Comment, 0)
	for _, parentID := range parentIDs {
		children := byParent[parentID]
		sort.Slice(children, func(i, j int) bool {
			if children[i].CreatedAt().Equal(children[j].CreatedAt()) {
				return children[i].ID().String() > children[j].ID().String()
			}
			return children[i].CreatedAt().After(children[j].CreatedAt())
		})

		if len(children) > limit {
			children = children[:limit]
		}

		result = append(result, children...)
	}
	return result, nil
}

func (comrepos *CommentRepos) GetRootsByPostIDs(ctx context.Context, postIDs []uuid.UUID, limit int) ([]*comment.Comment, error) {
	comrepos.mu.RLock()
	defer comrepos.mu.RUnlock()

	if len(postIDs) == 0 || limit <= 0 {
		return []*comment.Comment{}, nil
	}
	byPost := make(map[uuid.UUID][]*comment.Comment)
	for _, c := range comrepos.comments {
		if c.ParentID() == nil {
			for _, pid := range postIDs {
				if c.PostID() == pid {
					byPost[pid] = append(byPost[pid], c)
					break
				}
			}
		}
	}
	var result []*comment.Comment
	for _, list := range byPost {
		sortComments(list)
		if len(list) > limit {
			list = list[:limit]
		}
		result = append(result, list...)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].PostID() == result[j].PostID() {
			if result[i].CreatedAt().Equal(result[j].CreatedAt()) {
				return result[i].ID().String() > result[j].ID().String()
			}
			return result[i].CreatedAt().After(result[j].CreatedAt())
		}
		return result[i].PostID().String() < result[j].PostID().String()
	})
	return result, nil
}

func sortComments(comments []*comment.Comment) {
	sort.Slice(comments, func(i, j int) bool {
		if comments[i].CreatedAt().Equal(comments[j].CreatedAt()) {
			return comments[i].ID().String() > comments[j].ID().String()
		}

		return comments[i].CreatedAt().After(comments[j].CreatedAt())
	})
}
