-- +goose up

CREATE TABLE IF NOT EXISTS transactions(
    id INTEGER PRIMARY KEY,
    account_id INTEGER NOT NULL REFERENCES accounts(id),
    to_account_id INTEGER REFERENCES accounts(id),
    type TEXT NOT NULL CHECK(type IN ('income', 'expense', 'transfer')),
    amount INTEGER NOT NULL CHECK(amount > 0), -- minor units (cents)
    description TEXT NOT NULL DEFAULT '',
    date DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_transactions_account_date ON transactions(account_id, date DESC);

-- +goose down
DROP TABLE IF EXISTS transactions;
