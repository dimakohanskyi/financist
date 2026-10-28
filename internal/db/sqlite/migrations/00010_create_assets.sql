-- +goose up

-- Manually valued net-worth items. Cash isn't one: it comes from account balances.
CREATE TABLE IF NOT EXISTS assets(
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL CHECK(name <> ''),
    asset_class TEXT NOT NULL CHECK(asset_class IN ('investments', 'real_estate', 'vehicles', 'precious_metals', 'other')),
    currency TEXT NOT NULL DEFAULT 'USD',
    notes TEXT NOT NULL DEFAULT '',
    -- user-assigned: a laptop can earn (income), just exist (neutral) or cost money (drain)
    capital_role TEXT NOT NULL DEFAULT 'neutral' CHECK(capital_role IN ('income', 'neutral', 'drain')),
    monthly_cash_flow INTEGER, -- rough self-reported net per month, informational only
    risk_level TEXT NOT NULL DEFAULT 'medium' CHECK(risk_level IN ('low', 'medium', 'high'))
);

-- Point-in-time values; the latest is the current value.
CREATE TABLE IF NOT EXISTS asset_valuations(
    id INTEGER PRIMARY KEY,
    asset_id INTEGER NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    value INTEGER NOT NULL CHECK(value >= 0), -- minor units (cents)
    as_of_date DATETIME NOT NULL,
    UNIQUE(asset_id, as_of_date)
);

-- +goose down
DROP TABLE IF EXISTS asset_valuations;
DROP TABLE IF EXISTS assets;
