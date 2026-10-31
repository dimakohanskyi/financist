-- +goose up

-- What the destination account receives, in its own currency. Set only on
-- transfers between accounts of different currencies.
ALTER TABLE transactions ADD COLUMN to_amount INTEGER CHECK(to_amount IS NULL OR to_amount > 0); -- minor units (cents)

-- +goose down
ALTER TABLE transactions DROP COLUMN to_amount;
