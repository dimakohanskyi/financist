-- +goose up

-- One monthly limit per expense category, in effect until changed.
CREATE TABLE IF NOT EXISTS budgets(
    id INTEGER PRIMARY KEY,
    category_id INTEGER NOT NULL UNIQUE REFERENCES categories(id) ON DELETE CASCADE,
    monthly_limit INTEGER NOT NULL CHECK(monthly_limit > 0) -- minor units (cents)
);

-- +goose down
DROP TABLE IF EXISTS budgets;
