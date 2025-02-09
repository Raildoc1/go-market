BEGIN TRANSACTION;

CREATE TABLE withdrawals
(
    order_number VARCHAR(1024) PRIMARY KEY,
    user_id      INT       NOT NULL,
    amount       BIGINT    NOT NULL,
    process_time TIMESTAMP NOT NULL
);

COMMIT;