-- +goose up

-- A template for a transaction that repeats (rent, salary, subscriptions).
-- Posting is a manual action, so a missed week never back-fills a pile of rows.
CREATE TABLE IF NOT EXISTS recurring_transactions(
    id INTEGER PRIMARY KEY,
    account_id INTEGER NOT NULL REFERENCES accounts(id),
    to_account_id INTEGER REFERENCES accounts(id),
    category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
    type TEXT NOT NULL CHECK(type IN ('income', 'expense', 'transfer')),
    amount INTEGER NOT NULL CHECK(amount > 0), -- minor units (cents)
    description TEXT NOT NULL DEFAULT '',
    merchant TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    frequency TEXT NOT NULL CHECK(frequency IN ('weekly', 'monthly', 'yearly')),
    anchor_date DATETIME NOT NULL, -- first due date
    last_posted_date DATETIME,
    is_active BOOL NOT NULL DEFAULT TRUE -- paused templates keep their history
);

-- +goose down
DROP TABLE IF EXISTS recurring_transactions;
