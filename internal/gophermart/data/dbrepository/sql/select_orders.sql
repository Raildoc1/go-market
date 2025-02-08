SELECT number
FROM orders
WHERE status IN (%s)
LIMIT $1

