BEGIN TRANSACTION;

ALTER TABLE users
    ALTER id DROP DEFAULT;

DROP SEQUENCE users_id_seq;

ALTER TABLE users
    ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY;

SELECT setval(
               pg_get_serial_sequence('users', 'id'),
               coalesce((SELECT MAX(id) + 1 FROM users), 1),
               false
       );

COMMIT;