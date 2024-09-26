CREATE TABLE IF NOT EXISTS library (
    id SERIAL PRIMARY KEY,
    group_name TEXT NOT NULL,
    song TEXT NOT NULL,
    release_date DATE NOT NULL,
    text TEXT NOT NULL,
    patronymic TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_library_id ON library(id);
CREATE INDEX IF NOT EXISTS idx_library_group_name ON library(group_name);
CREATE INDEX IF NOT EXISTS idx_library_song ON library(song);
CREATE INDEX IF NOT EXISTS idx_library_release_date ON library(release_date);
CREATE INDEX IF NOT EXISTS idx_library_text ON library(text);