CREATE TABLE todos (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    text TEXT NOT NULL CHECK (LENGTH(TRIM(text)) > 0 AND LENGTH(text) <= 500),
    completed BOOLEAN NOT NULL DEFAULT false,
    due_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_todos_user_created_at ON todos(user_id, created_at DESC);
