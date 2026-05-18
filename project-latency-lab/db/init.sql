CREATE TABLE IF NOT EXISTS users (
    id           SERIAL PRIMARY KEY,
    username     VARCHAR(100) UNIQUE NOT NULL,
    token        VARCHAR(64) UNIQUE NOT NULL,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_users_token ON users(token);

CREATE TABLE IF NOT EXISTS notes (
    id           VARCHAR(20) PRIMARY KEY,
    author       VARCHAR(100) NOT NULL,
    body         TEXT NOT NULL,
    view_count   BIGINT DEFAULT 0,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_notes_author     ON notes(author);
CREATE INDEX IF NOT EXISTS idx_notes_created_at ON notes(created_at DESC);

CREATE TABLE IF NOT EXISTS views (
    id         SERIAL PRIMARY KEY,
    note_id    VARCHAR(20) NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    viewed_at  TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_views_note_id   ON views(note_id);
CREATE INDEX IF NOT EXISTS idx_views_viewed_at ON views(viewed_at);
