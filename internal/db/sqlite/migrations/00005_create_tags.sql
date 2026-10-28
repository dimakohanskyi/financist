-- +goose up

-- ponytail: NOCASE folds ASCII only, so "Київ"/"київ" stay two tags; add an ICU collation if that bites
CREATE TABLE IF NOT EXISTS tags(
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE COLLATE NOCASE CHECK(name <> '')
);

CREATE TABLE IF NOT EXISTS transaction_tags(
    transaction_id INTEGER NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY(transaction_id, tag_id)
);
CREATE INDEX IF NOT EXISTS idx_transaction_tags_tag ON transaction_tags(tag_id);

ALTER TABLE transactions ADD COLUMN merchant TEXT NOT NULL DEFAULT '';
ALTER TABLE transactions ADD COLUMN notes TEXT NOT NULL DEFAULT '';

-- +goose down
ALTER TABLE transactions DROP COLUMN notes;
ALTER TABLE transactions DROP COLUMN merchant;
DROP TABLE IF EXISTS transaction_tags;
DROP TABLE IF EXISTS tags;
