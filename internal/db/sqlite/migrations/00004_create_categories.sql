-- +goose up

CREATE TABLE IF NOT EXISTS categories(
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL CHECK(name <> ''),
    kind TEXT NOT NULL CHECK(kind IN ('income', 'expense')),
    icon TEXT, -- lucide icon name
    color TEXT NOT NULL CHECK(color GLOB '#[0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f]'),
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_default BOOL NOT NULL DEFAULT FALSE,
    -- one level deep only, enforced by the service; deleting a parent makes its children top-level
    parent_id INTEGER REFERENCES categories(id) ON DELETE SET NULL
);

INSERT INTO categories(name, kind, icon, color, sort_order, is_default) VALUES
    ('Housing & Utilities', 'expense', 'home', '#2a78d6', 0, TRUE),
    ('Groceries', 'expense', 'shopping-basket', '#1baf7a', 1, TRUE),
    ('Dining Out', 'expense', 'utensils', '#eb6834', 2, TRUE),
    ('Transportation', 'expense', 'car', '#4a3aa7', 3, TRUE),
    ('Health & Fitness', 'expense', 'heart-pulse', '#e34948', 4, TRUE),
    ('Shopping', 'expense', 'shopping-bag', '#eda100', 5, TRUE),
    ('Entertainment', 'expense', 'clapperboard', '#e87ba4', 6, TRUE),
    ('Subscriptions', 'expense', 'repeat', '#008300', 7, TRUE),
    ('Salary', 'income', 'banknote', '#2a78d6', 8, TRUE),
    ('Freelance', 'income', 'briefcase', '#1baf7a', 9, TRUE),
    ('Investments', 'income', 'trending-up', '#4a3aa7', 10, TRUE),
    ('Gifts', 'income', 'gift', '#e87ba4', 11, TRUE),
    ('Business Income', 'income', 'building-2', '#eb6834', 12, TRUE),
    ('Rental Income', 'income', 'key', '#eda100', 13, TRUE),
    ('Benefits', 'income', 'hand-coins', '#008300', 14, TRUE),
    ('Item Sales', 'income', 'tag', '#e34948', 15, TRUE),
    ('Other Income', 'income', 'plus-circle', '#898781', 16, TRUE);

ALTER TABLE transactions ADD COLUMN category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_category ON transactions(category_id);

-- +goose down

-- SQLite can't drop a foreign key column, so transactions is rebuilt without it.
CREATE TABLE transactions_old(
    id INTEGER PRIMARY KEY,
    account_id INTEGER NOT NULL REFERENCES accounts(id),
    to_account_id INTEGER REFERENCES accounts(id),
    type TEXT NOT NULL CHECK(type IN ('income', 'expense', 'transfer')),
    amount INTEGER NOT NULL CHECK(amount > 0),
    description TEXT NOT NULL DEFAULT '',
    date DATETIME NOT NULL
);
INSERT INTO transactions_old SELECT id, account_id, to_account_id, type, amount, description, date FROM transactions;
DROP TABLE transactions;
ALTER TABLE transactions_old RENAME TO transactions;
CREATE INDEX IF NOT EXISTS idx_transactions_account_date ON transactions(account_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_to_account ON transactions(to_account_id);
DROP TABLE IF EXISTS categories;
