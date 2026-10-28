-- +goose up

-- Single row (id 1): display currency and the thresholds behind alerts.
CREATE TABLE IF NOT EXISTS app_settings(
    id INTEGER PRIMARY KEY CHECK(id = 1),
    currency TEXT NOT NULL DEFAULT 'USD',
    -- complete months in a row before the alert fires
    negative_cash_flow_threshold_months INTEGER NOT NULL DEFAULT 2 CHECK(negative_cash_flow_threshold_months BETWEEN 1 AND 24),
    net_worth_decline_threshold_months INTEGER NOT NULL DEFAULT 2 CHECK(net_worth_decline_threshold_months BETWEEN 1 AND 24),
    -- max % of capital in medium/high risk, the "20" of an 80/20 rule
    risky_allocation_threshold_percent INTEGER NOT NULL DEFAULT 20 CHECK(risky_allocation_threshold_percent BETWEEN 1 AND 100),
    -- a cash account at or above this balance, untouched for this many days, is idle
    idle_cash_threshold_amount INTEGER NOT NULL DEFAULT 100000 CHECK(idle_cash_threshold_amount >= 0), -- minor units (cents)
    idle_cash_threshold_days INTEGER NOT NULL DEFAULT 60 CHECK(idle_cash_threshold_days BETWEEN 1 AND 365)
);
INSERT INTO app_settings(id) VALUES (1);

-- +goose down
DROP TABLE IF EXISTS app_settings;
