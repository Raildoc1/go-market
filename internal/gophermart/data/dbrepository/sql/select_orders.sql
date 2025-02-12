SELECT number, accrual, upload_time, status
FROM orders
WHERE user_id = $1

