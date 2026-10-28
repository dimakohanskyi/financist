-- +goose up

CREATE TABLE IF NOT EXISTS goals(
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL CHECK(name <> ''),
    target_amount INTEGER NOT NULL CHECK(target_amount > 0), -- minor units (cents)
    target_date DATETIME
);

-- A goal's current amount is the sum of its contributions; negative ones are withdrawals.
CREATE TABLE IF NOT EXISTS goal_contributions(
    id INTEGER PRIMARY KEY,
    goal_id INTEGER NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    amount INTEGER NOT NULL CHECK(amount <> 0),
    date DATETIME NOT NULL,
    note TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_goal_contributions_goal ON goal_contributions(goal_id);

-- +goose down
DROP TABLE IF EXISTS goal_contributions;
DROP TABLE IF EXISTS goals;
