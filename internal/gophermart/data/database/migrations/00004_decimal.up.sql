BEGIN TRANSACTION;

-- ALTER TABLE withdrawals
-- ADD COLUMN amount_tmp DECIMAL(16, 3);
--
-- UPDATE withdrawals
-- SET amount_tmp = cast(amount as DECIMAL(16, 3));
--
-- ALTER TABLE withdrawals
-- DROP COLUMN amount;

ALTER TABLE withdrawals
    ALTER COLUMN amount TYPE DECIMAL(16, 3);

ALTER TABLE users
    ALTER COLUMN balance TYPE DECIMAL(16, 3);

ALTER TABLE orders
    ALTER COLUMN accrual TYPE DECIMAL(16, 3);

COMMIT;