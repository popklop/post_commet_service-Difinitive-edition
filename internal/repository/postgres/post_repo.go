package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"ozon-testovoe/internal/entity/post"
	"time"

	"github.com/google/uuid"
)

type PostRepos struct {
	db *sql.DB
}

func NewPostRepos(db *sql.DB) *PostRepos {
	return &PostRepos{db: db}
}

func (r *PostRepos) Create(ctx context.Context, p *post.Post) error {
	query := `
		INSERT INTO posts (id, title, content, author, comments_enabled, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		p.ID(),
		p.Title(),
		p.Content(),
		p.Author(),
		p.CommentsEnabled(),
		p.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}
	return nil
}

func (r *PostRepos) GetById(ctx context.Context, id uuid.UUID) (*post.Post, error) {
	var postID uuid.UUID
	var title string
	var content string
	var author uuid.UUID
	var commentsEnabled bool
	var createdAt time.Time

	query := `SELECT id, title, content, author, comments_enabled, created_at FROM posts WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&postID, &title, &content, &author, &commentsEnabled, &createdAt)
	if err == sql.ErrNoRows {
		return nil, ErrPostNotFound
	}
	if err != nil {
		return nil, err
	}

	return post.NewPostFromDB(postID, title, content, author, commentsEnabled, createdAt), nil
}

func (postrepos *PostRepos) List(ctx context.Context, first int, after *string) ([]*post.Post, *string, bool, error) {
	if first <= 0 {
		return nil, nil, false, ErrNegativeFirst
	}

	var afterID uuid.UUID
	var afterCreatedAt time.Time
	hasCursor := after != nil

	if hasCursor {
		var err error
		afterID, err = uuid.Parse(*after)
		if err != nil {
			return nil, nil, false, ErrInvalidCursor
		}
		err = postrepos.db.QueryRowContext(ctx, "SELECT created_at FROM posts WHERE id = $1", afterID).Scan(&afterCreatedAt)
		if err == sql.ErrNoRows {
			return nil, nil, false, ErrCursorNotFound
		}
		if err != nil {
			return nil, nil, false, fmt.Errorf("failed to get cursor: %w", err)
		}
	}

	query := `
		SELECT id, title, content, author, comments_enabled, created_at
		FROM posts
	`
	args := []interface{}{}
	if hasCursor {
		query += ` WHERE (created_at, id) < ($1, $2)`
		args = append(args, afterCreatedAt, afterID)
	}
	query += ` ORDER BY created_at DESC, id DESC LIMIT $` + fmt.Sprint(len(args)+1)
	args = append(args, first+1)

	rows, err := postrepos.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to query posts: %w", err)
	}
	defer rows.Close()

	posts := make([]*post.Post, 0, first+1)
	for rows.Next() {
		var pID uuid.UUID
		var title, content string
		var author uuid.UUID
		var commentsEnabled bool
		var createdAt time.Time

		if err := rows.Scan(&pID, &title, &content, &author, &commentsEnabled, &createdAt); err != nil {
			return nil, nil, false, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, post.NewPostFromDB(pID, title, content, author, commentsEnabled, createdAt))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, false, fmt.Errorf("rows iteration error: %w", err)
	}

	haveWeMoreAfter := len(posts) > first
	if haveWeMoreAfter {
		posts = posts[:first]
	}

	var endCursor *string
	if len(posts) > 0 {
		cursor := posts[len(posts)-1].ID().String()
		endCursor = &cursor
	}

	return posts, endCursor, haveWeMoreAfter, nil
}

func (r *PostRepos) UpdateCommentsEnabled(ctx context.Context, id uuid.UUID, enabled bool) error {
	var postIdFromDB string
	query := `UPDATE posts
	set comments_enabled = $1 WHERE id = $2 RETURNING id`
	err := r.db.QueryRowContext(ctx, query, enabled, id).Scan(&postIdFromDB)
	if err == sql.ErrNoRows {
		return ErrPostNotOnUpdate
	}
	if err != nil {
		return err
	}
	return nil
}
