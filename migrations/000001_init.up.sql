CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE posts (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    author UUID NOT NULL,
    comments_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);
CREATE TABLE comments (
    id UUID PRIMARY KEY,
    text TEXT NOT NULL,
    author UUID NOT NULL,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    parent_id UUID NULL REFERENCES comments(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_posts_created_at ON posts(created_at DESC);
CREATE INDEX idx_comments_post_id ON comments(post_id);
CREATE INDEX idx_comments_parent_id ON comments(parent_id);
CREATE INDEX idx_comments_created_at ON comments(created_at DESC);
CREATE INDEX idx_posts_created_at_id ON posts(created_at DESC, id DESC);
CREATE INDEX idx_comments_post_id_created_at_id ON comments(post_id, created_at DESC, id DESC);
CREATE INDEX idx_comments_parent_id_created_at_id ON comments(parent_id, created_at DESC, id DESC);