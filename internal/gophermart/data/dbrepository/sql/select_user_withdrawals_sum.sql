SELECT SUM(amount)
FROM withdrawals
WHERE user_id = $1