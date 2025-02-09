SHOW INDEX FROM profiles;
SHOW INDEX FROM permissions;
SHOW INDEX FROM users;
SHOW INDEX FROM comments;

DELETE FROM users WHERE users.id = 2;

SELECT * FROM access_tokens WHERE status = 1;

DESCRIBE profiles;

SELECT * FROM users;
SELECT * FROM roles;
SELECT * FROM permissions;
SELECT * FROM profiles INNER JOIN users ON profiles.user_id = users.id WHERE users.id = 2;

DELETE FROM users;
SELECT * FROM `users` INNER JOIN profiles ON profiles.user_id = users.id WHERE users.id = 2 ORDER BY `users`.`id` LIMIT 1;


-- Authentication and Authorization setup
INSERT INTO roles (id, name, created_at, updated_at)
VALUES (1, 'admin', now(), now()),
       (2, 'user', now(), now());

UPDATE users SET role_id = 2 WHERE id = 1;
INSERT INTO permissions (id, name, created_at, updated_at)
VALUES (1,  'allow_create_role',now(), now()),
       (2, 'allow_create_permission',now(), now());

INSERT INTO role_permissions (id, role_id, permission_id, created_at, updated_at)
VALUES (1, 1, 1, now(), now()),
       (2, 1, 2, now(),now());

-- Continue to add some more permission to assign by postman
-- ["allow_assign_permissions", "allow_list_roles", "allow_list_permissions"] and then execute sql below
INSERT INTO role_permissions (id, role_id, permission_id, created_at, updated_at)
VALUES (3, 1, 3, now(), now()),
       (4, 1, 4, now(), now()),
       (5, 1, 5, now(), now());

SELECT * FROM role_permissions;