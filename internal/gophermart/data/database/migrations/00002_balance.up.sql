BEGIN TRANSACTION;

ALTER TABLE users
    ADD COLUMN balance BIGINT;

WITH accruals AS (SELECT SUM(accrual) as total_accruals, user_id
                  FROM orders
                  GROUP BY user_id)
UPDATE users
SET balance = accruals.total_accruals
FROM accruals
WHERE id = accruals.user_id;

COMMIT;