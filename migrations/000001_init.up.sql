CREATE TABLE posts (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    author_id BIGINT NOT NULL,
    comments_allowed BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    content TEXT NOT NULL CHECK (char_length(content) BETWEEN 1 AND 2000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    parent_comment_id BIGINT,
    post_id INT REFERENCES posts(id) ON DELETE CASCADE NOT NULL,
    author_id BIGINT NOT NULL,

    CONSTRAINT uq_comment_id_post UNIQUE (id, post_id),

    CONSTRAINT fk_comment_parent FOREIGN KEY (parent_comment_id, post_id) REFERENCES comments(id, post_id)
);

CREATE INDEX idx_comments_parent_created_id ON comments(parent_comment_id, created_at, id);

CREATE INDEX idx_comments_roots ON comments(post_id, created_at, id)
    WHERE parent_comment_id IS NULL;
