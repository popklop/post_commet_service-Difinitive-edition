package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"ozon-testovoe/internal/entity/comment"
	"strings"
	"time"

	"github.com/google/uuid"
)

type CommentRepos struct {
	db *sql.DB
}

func NewCommentRepos(db *sql.DB) *CommentRepos {
	return &CommentRepos{db: db}
}

func (comrepos *CommentRepos) Create(ctx context.Context, c *comment.Comment) error {
	query := `
		INSERT INTO comments (id, text, author, post_id, parent_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	var parentID *uuid.UUID
	if c.ParentID() != nil {
		pid := c.ParentID()
		parentID = pid
	}
	_, err := comrepos.db.ExecContext(ctx, query,
		c.ID(),
		c.Text(),
		c.Author(),
		c.PostID(),
		parentID,
		c.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}
	return nil
}

func (comrepos *CommentRepos) GetByID(ctx context.Context, id uuid.UUID) (*comment.Comment, error) {
	var cID uuid.UUID
	var text string
	var author uuid.UUID
	var postID uuid.UUID
	var parentID *uuid.UUID
	var createdAt time.Time

	query := `SELECT id, text, author, post_id, parent_id, created_at FROM comments WHERE id = $1`
	err := comrepos.db.QueryRowContext(ctx, query, id).Scan(&cID, &text, &author, &postID, &parentID, &createdAt)
	if err == sql.ErrNoRows {
		return nil, ErrCommNotFound
	}
	if err != nil {
		return nil, err
	}
	return comment.NewCommentFromDB(cID, text, author, postID, parentID, createdAt), nil
}

func (comrepos *CommentRepos) GetRootComments(ctx context.Context, postID uuid.UUID, first int, after *string) ([]*comment.Comment, *string, bool, error) {
	if first <= 0 {
		return nil, nil, false, ErrNegativeFirst
	}

	var cursorID uuid.UUID
	var cursorCreatedAt time.Time
	hasCursor := after != nil
	if hasCursor {
		var err error
		cursorID, err = uuid.Parse(*after)
		if err != nil {
			return nil, nil, false, ErrInvalidCursor
		}
		err = comrepos.db.QueryRowContext(ctx, "SELECT created_at FROM comments WHERE id = $1", cursorID).Scan(&cursorCreatedAt)
		if err == sql.ErrNoRows {
			return nil, nil, false, ErrCursorNotFound
		}
		if err != nil {
			return nil, nil, false, fmt.Errorf("failed to get cursor: %w", err)
		}
	}

	query := `
		SELECT id, text, author, post_id, parent_id, created_at
		FROM comments
		WHERE post_id = $1 AND parent_id IS NULL
	`
	args := []interface{}{postID}
	if hasCursor {
		query += ` AND (created_at, id) < ($2, $3)`
		args = append(args, cursorCreatedAt, cursorID)
	}
	query += ` ORDER BY created_at DESC, id DESC LIMIT $` + fmt.Sprint(len(args)+1)
	args = append(args, first+1)

	rows, err := comrepos.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to query root comments: %w", err)
	}
	defer rows.Close()

	rootcomments := make([]*comment.Comment, 0, first+1)
	for rows.Next() {
		var cID uuid.UUID
		var text string
		var author uuid.UUID
		var pID uuid.UUID
		var parentID *uuid.UUID
		var createdAt time.Time

		if err := rows.Scan(&cID, &text, &author, &pID, &parentID, &createdAt); err != nil {
			return nil, nil, false, fmt.Errorf("failed to scan comment: %w", err)
		}
		rootcomments = append(rootcomments, comment.NewCommentFromDB(cID, text, author, pID, parentID, createdAt))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, false, fmt.Errorf("rows iteration error: %w", err)
	}

	haveWeMoreAfter := len(rootcomments) > first
	if haveWeMoreAfter {
		rootcomments = rootcomments[:first]
	}
	var lastused *string
	if len(rootcomments) > 0 {
		cursor := rootcomments[len(rootcomments)-1].ID().String()
		lastused = &cursor
	}

	return rootcomments, lastused, haveWeMoreAfter, nil
}

func (comrepos *CommentRepos) GetChildComments(ctx context.Context, parentID uuid.UUID, first int, after *string) ([]*comment.Comment, *string, bool, error) {
	if first <= 0 {
		return nil, nil, false, ErrNegativeFirst
	}

	var cursorID uuid.UUID
	var cursorCreatedAt time.Time
	hasCursor := after != nil
	if hasCursor {
		var err error
		cursorID, err = uuid.Parse(*after)
		if err != nil {
			return nil, nil, false, ErrInvalidCursor
		}
		err = comrepos.db.QueryRowContext(ctx, "SELECT created_at FROM comments WHERE id = $1", cursorID).Scan(&cursorCreatedAt)
		if err == sql.ErrNoRows {
			return nil, nil, false, ErrCursorNotFound
		}
		if err != nil {
			return nil, nil, false, fmt.Errorf("failed to get cursor: %w", err)
		}
	}

	query := `
		SELECT id, text, author, post_id, parent_id, created_at
		FROM comments
		WHERE parent_id = $1`
	args := []interface{}{parentID}
	if hasCursor {
		query += ` AND (created_at, id) < ($2, $3)`
		args = append(args, cursorCreatedAt, cursorID)
	}
	query += ` ORDER BY created_at DESC, id DESC LIMIT $` + fmt.Sprint(len(args)+1)
	args = append(args, first+1)

	rows, err := comrepos.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to query child comments: %w", err)
	}
	defer rows.Close()

	childcomments := make([]*comment.Comment, 0, first+1)
	for rows.Next() {
		var cID uuid.UUID
		var text string
		var author uuid.UUID
		var pID uuid.UUID
		var parentIDPtr *uuid.UUID
		var createdAt time.Time

		if err := rows.Scan(&cID, &text, &author, &pID, &parentIDPtr, &createdAt); err != nil {
			return nil, nil, false, fmt.Errorf("failed to scan comment: %w", err)
		}
		childcomments = append(childcomments, comment.NewCommentFromDB(cID, text, author, pID, parentIDPtr, createdAt))
	}
	if err := rows.Err(); err != nil {
		return nil, nil, false, fmt.Errorf("rows iteration error: %w", err)
	}

	haveWeMoreAfter := len(childcomments) > first
	if haveWeMoreAfter {
		childcomments = childcomments[:first]
	}

	var lastfound *string
	if len(childcomments) > 0 {
		cursor := childcomments[len(childcomments)-1].ID().String()
		lastfound = &cursor
	}

	return childcomments, lastfound, haveWeMoreAfter, nil
}

func (comrepos *CommentRepos) GetByParentIDs(ctx context.Context, parentIDs []uuid.UUID, limit int) ([]*comment.Comment, error) {
	if len(parentIDs) == 0 || limit <= 0 {
		return []*comment.Comment{}, nil
	}

	placeholders := make([]string, len(parentIDs))
	args := make([]interface{}, len(parentIDs))
	for i, id := range parentIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	limitArg := len(parentIDs) + 1
	args = append(args, limit)

	query := fmt.Sprintf(`
		WITH ranked AS (
			SELECT
				id,
				text,
				author,
				post_id,
				parent_id,
				created_at,
				ROW_NUMBER() OVER (
					PARTITION BY parent_id
					ORDER BY created_at DESC, id DESC
				) AS rn
			FROM comments
			WHERE parent_id IN (%s)
		)
		SELECT
			id,
			text,
			author,
			post_id,
			parent_id,
			created_at
		FROM ranked
		WHERE rn <= $%d
		ORDER BY parent_id, created_at DESC, id DESC
	`, strings.Join(placeholders, ", "), limitArg)

	rows, err := comrepos.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments by parent IDs: %w", err)
	}
	defer rows.Close()

	comments := make([]*comment.Comment, 0)
	for rows.Next() {
		var cID uuid.UUID
		var text string
		var author uuid.UUID
		var postID uuid.UUID
		var parentID *uuid.UUID
		var createdAt time.Time

		if err := rows.Scan(&cID, &text, &author, &postID, &parentID, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment.NewCommentFromDB(cID, text, author, postID, parentID, createdAt))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return comments, nil
}

func (comrepos *CommentRepos) GetRootsByPostIDs(ctx context.Context, postIDs []uuid.UUID, limit int) ([]*comment.Comment, error) {
	if len(postIDs) == 0 || limit <= 0 {
		return []*comment.Comment{}, nil
	}

	placeholders := make([]string, len(postIDs))
	args := make([]interface{}, len(postIDs))
	for i, id := range postIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
        WITH ranked AS (
            SELECT id, text, author, post_id, parent_id, created_at,
                   ROW_NUMBER() OVER (PARTITION BY post_id ORDER BY created_at DESC, id DESC) as rn
            FROM comments
            WHERE parent_id IS NULL AND post_id IN (%s)
        )
        SELECT id, text, author, post_id, parent_id, created_at
        FROM ranked
        WHERE rn <= $%d
        ORDER BY post_id, created_at DESC, id DESC
    `, strings.Join(placeholders, ", "), len(postIDs)+1)
	args = append(args, limit)

	rows, err := comrepos.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query root comments by post IDs: %w", err)
	}
	defer rows.Close()

	var comments []*comment.Comment
	for rows.Next() {
		var cID uuid.UUID
		var text string
		var author uuid.UUID
		var postID uuid.UUID
		var parentID *uuid.UUID
		var createdAt time.Time

		if err := rows.Scan(&cID, &text, &author, &postID, &parentID, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment.NewCommentFromDB(cID, text, author, postID, parentID, createdAt))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return comments, nil
}
