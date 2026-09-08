package dataloader

import (
	"context"
	"time"

	"ozon-testovoe/internal/service"

	"github.com/google/uuid"
	gophersdataloader "github.com/graph-gophers/dataloader/v7"
)

type Loaders struct {
	CommentChildrenLoader *gophersdataloader.Loader[string, interface{}]
	PostCommentsLoader    *gophersdataloader.Loader[string, interface{}]
}

func NewLoaders(commentService service.CommentService) *Loaders {
	const childrenLimit = 20

	commentChildrenLoader := gophersdataloader.NewBatchedLoader[string, interface{}](
		func(ctx context.Context, keys []string) []*gophersdataloader.Result[interface{}] {
			results := make([]*gophersdataloader.Result[interface{}], len(keys))

			parentIDs := make([]uuid.UUID, 0, len(keys))
			validIndexes := make(map[uuid.UUID][]int, len(keys))

			for i, key := range keys {
				parentID, err := uuid.Parse(key)
				if err != nil {
					results[i] = &gophersdataloader.Result[interface{}]{Error: err}
					continue
				}
				parentIDs = append(parentIDs, parentID)
				validIndexes[parentID] = append(validIndexes[parentID], i)
			}

			if len(parentIDs) == 0 {
				return results
			}

			childrenMap, err := commentService.GetChildrenBatch(ctx, parentIDs, childrenLimit)
			if err != nil {
				for i := range keys {
					if results[i] == nil {
						results[i] = &gophersdataloader.Result[interface{}]{Error: err}
					}
				}
				return results
			}

			for parentID, indexes := range validIndexes {
				data := childrenMap[parentID]
				for _, i := range indexes {
					results[i] = &gophersdataloader.Result[interface{}]{Data: data}
				}
			}
			return results
		},
		gophersdataloader.WithWait[string, interface{}](2*time.Millisecond),
	)

	postCommentsLoader := gophersdataloader.NewBatchedLoader[string, interface{}](
		func(ctx context.Context, keys []string) []*gophersdataloader.Result[interface{}] {
			results := make([]*gophersdataloader.Result[interface{}], len(keys))

			postIDs := make([]uuid.UUID, 0, len(keys))
			validIndexes := make(map[uuid.UUID][]int, len(keys))

			for i, key := range keys {
				postID, err := uuid.Parse(key)
				if err != nil {
					results[i] = &gophersdataloader.Result[interface{}]{Error: err}
					continue
				}
				postIDs = append(postIDs, postID)
				validIndexes[postID] = append(validIndexes[postID], i)
			}

			if len(postIDs) == 0 {
				return results
			}

			commentsMap, err := commentService.GetRootsBatch(ctx, postIDs)
			if err != nil {
				for i := range keys {
					if results[i] == nil {
						results[i] = &gophersdataloader.Result[interface{}]{Error: err}
					}
				}
				return results
			}

			for postID, indexes := range validIndexes {
				data := commentsMap[postID]
				for _, i := range indexes {
					results[i] = &gophersdataloader.Result[interface{}]{Data: data}
				}
			}
			return results
		},
		gophersdataloader.WithWait[string, interface{}](2*time.Millisecond),
	)

	return &Loaders{
		CommentChildrenLoader: commentChildrenLoader,
		PostCommentsLoader:    postCommentsLoader,
	}
}

type contextKey struct{}

var loadersKey contextKey

func WithLoaders(ctx context.Context, loaders *Loaders) context.Context {
	return context.WithValue(ctx, loadersKey, loaders)
}

func For(ctx context.Context) *Loaders {
	loaders, ok := ctx.Value(loadersKey).(*Loaders)
	if !ok {
		panic("dataloader: loaders not found in context")
	}
	return loaders
}
