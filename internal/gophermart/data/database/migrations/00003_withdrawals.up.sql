BEGIN TRANSACTION;

CREATE TABLE withdrawals
(
    id           SERIAL PRIMARY KEY,
    order_number VARCHAR(1024),
    amount       BIGINT NOT NULL,
    process_time TIMESTAMP
);

COMMIT;