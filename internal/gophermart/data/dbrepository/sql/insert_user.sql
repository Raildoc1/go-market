INSERT INTO users (login, password)
VALUES ($1, crypt($2, gen_salt('md5')))
RETURNING id