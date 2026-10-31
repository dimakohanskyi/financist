-- +goose up

-- What was actually paid when it was in a currency other than the account's;
-- amount then holds the converted value in the account's currency.
ALTER TABLE transactions ADD COLUMN original_amount INTEGER CHECK(original_amount IS NULL OR original_amount > 0); -- minor units (cents)
ALTER TABLE transactions ADD COLUMN original_currency TEXT;

-- +goose down
ALTER TABLE transactions DROP COLUMN original_currency;
ALTER TABLE transactions DROP COLUMN original_amount;
