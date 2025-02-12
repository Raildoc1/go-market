SELECT order_number, amount, process_time
FROM withdrawals
WHERE user_id = $1

