-- +goose up

CREATE TABLE IF NOT EXISTS accounts(
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL, 
    account_type TEXT NOT NULL CHECK(account_type IN ('checking', 'savings', 'credit_card', 'cash', 'investment', 'other')),
    currency TEXT DEFAULT 'USD' NOT NULL,
    is_archived BOOL DEFAULT FALSE
);

-- +goose down
DROP TABLE IF EXISTS accounts;