-- +goose up

-- Units of quote currency per 1 USD, one row per quote and day the source published.
CREATE TABLE IF NOT EXISTS fx_rates(
    date DATE NOT NULL,
    quote TEXT NOT NULL,
    rate REAL NOT NULL CHECK(rate > 0),
    fetched_at DATETIME NOT NULL,
    UNIQUE(date, quote)
);

-- +goose down
DROP TABLE IF EXISTS fx_rates;
