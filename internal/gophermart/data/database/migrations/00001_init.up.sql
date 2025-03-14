BEGIN TRANSACTION;

CREATE EXTENSION pgcrypto;

CREATE TABLE users
(
    id       SERIAL PRIMARY KEY,
    login    VARCHAR(32)  NOT NULL UNIQUE,
    password VARCHAR(128) NOT NULL
);

CREATE TABLE orders
(
    number      VARCHAR(1024) PRIMARY KEY,
    status      VARCHAR(32),
    user_id     INTEGER,
    upload_time TIMESTAMP,
    accrual     BIGINT
);

COMMIT;