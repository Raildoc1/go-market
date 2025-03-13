UPDATE orders
SET status = $2, accrual = $3
WHERE number = $1