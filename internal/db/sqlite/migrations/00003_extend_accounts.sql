-- +goose up

-- SQLite can't alter a CHECK constraint, so the table is rebuilt (migrations run with foreign keys off).
CREATE TABLE accounts_new(
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    account_type TEXT NOT NULL CHECK(account_type IN ('checking', 'debit_card', 'savings', 'credit_card', 'cash', 'investment', 'other')),
    currency TEXT DEFAULT 'USD' NOT NULL,
    color TEXT CHECK(color GLOB '#[0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f]'),
    is_archived BOOL DEFAULT FALSE
);
INSERT INTO accounts_new(id, name, account_type, currency, is_archived)
    SELECT id, name, account_type, currency, is_archived FROM accounts;
DROP TABLE accounts;
ALTER TABLE accounts_new RENAME TO accounts;

-- balances look transfers up by destination too
CREATE INDEX IF NOT EXISTS idx_transactions_to_account ON transactions(to_account_id);

-- +goose down
DROP INDEX IF EXISTS idx_transactions_to_account;

CREATE TABLE accounts_old(
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    account_type TEXT NOT NULL CHECK(account_type IN ('checking', 'savings', 'credit_card', 'cash', 'investment', 'other')),
    currency TEXT DEFAULT 'USD' NOT NULL,
    is_archived BOOL DEFAULT FALSE
);
INSERT INTO accounts_old(id, name, account_type, currency, is_archived)
    SELECT id, name, CASE account_type WHEN 'debit_card' THEN 'checking' ELSE account_type END, currency, is_archived FROM accounts;
DROP TABLE accounts;
ALTER TABLE accounts_old RENAME TO accounts;
