SELECT id,
       (password = crypt($2, password))
           AS password_match
FROM users
WHERE login = $1