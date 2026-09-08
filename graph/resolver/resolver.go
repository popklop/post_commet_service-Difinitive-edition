package resolver

import "ozon-testovoe/internal/service"

type Resolver struct {
	PostService    service.PostService
	CommentService service.CommentService
}
