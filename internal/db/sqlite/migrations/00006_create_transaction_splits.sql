-- +goose up

-- A split transaction has no category of its own; its amount is divided across these lines.
CREATE TABLE IF NOT EXISTS transaction_splits(
    id INTEGER PRIMARY KEY,
    transaction_id INTEGER NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    -- SET NULL so deleting a category never breaks reading old splits
    category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
    amount INTEGER NOT NULL CHECK(amount > 0), -- minor units (cents)
    note TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_transaction_splits_transaction ON transaction_splits(transaction_id);
CREATE INDEX IF NOT EXISTS idx_transaction_splits_category ON transaction_splits(category_id);

-- +goose down
DROP TABLE IF EXISTS transaction_splits;
