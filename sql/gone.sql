-- Basic setup for new system
-- Create Categories
INSERT INTO categories (name) VALUES
                                  ('Social Media Trends'),
                                  ('Anime Fan Communities'),
                                  ('Music Production Tips'),
                                  ('Programming Tutorials'),
                                  ('Digital Art & Design'),
                                  ('K-Pop Culture'),
                                  ('Web Development'),
                                  ('Cosplay Inspiration'),
                                  ('Game Development'),
                                  ('AI & Machine Learning'),
                                  ('Manga Discussions'),
                                  ('Indie Music Artists'),
                                  ('Cybersecurity Corner'),
                                  ('Live Streaming Tips'),
                                  ('Anime Reviews'),
                                  ('Hip-Hop Culture'),
                                  ('Mobile App Development'),
                                  ('Viral Content Analysis'),
                                  ('Tech Gadgets Talk'),
                                  ('Songwriting & Composition');
-- Authentication and Authorization setup
INSERT INTO roles (id, name, created_at, updated_at)
VALUES (1, 'admin', now(), now()),
       (2, 'user', now(), now());

-- Try to log in or register at frontend, role_id need set to 1
UPDATE users SET role_id = 2 WHERE id = 3;
INSERT INTO permissions (id, name, created_at, updated_at)
VALUES (1,  'allow_create_role',now(), now()),
       (2, 'allow_create_permission',now(), now());

INSERT INTO role_permissions (id, role_id, permission_id, created_at, updated_at)
VALUES (1, 1,    1, now(), now()),
       (2, 1, 2, now(),now());

-- Continue to add some more permissions to assign by postman
-- ["allow_assign_permissions", "allow_list_roles", "allow_list_permissions"] and then execute sql below
INSERT INTO role_permissions (id, role_id, permission_id, created_at, updated_at)
VALUES (3, 1, 3, now(), now()),
       (4, 1, 4, now(), now()),
       (5, 1, 5, now(), now());


-- Checking block
SELECT * FROM users;
SELECT * FROM role_permissions;
SELECT * FROM access_tokens WHERE status = 1;
SELECT * FROM media;
SELECT * FROM posts WHERE id = 2;

SELECT posts.id, posts.content, posts.created_at, posts.updated_at,
media.id AS media_id, media.path AS media_path, media.type AS media_type, media.status AS media_status,
       categories.id AS categories_id, categories.name AS categories_name
FROM `posts`
    LEFT JOIN media ON media.post_id = posts.id
    LEFT JOIN post_categories ON post_categories.post_id = posts.id
    LEFT JOIN categories ON categories.id = post_categories.category_id
WHERE LOWER(posts.content) LIKE LOWER('%%') ORDER BY posts.created_at desc LIMIT 10;

SELECT posts.*, media.* FROM posts INNER JOIN media ON posts.id = media.post_id
                        WHERE posts.content LIKE '%%' ORDER BY posts.created_at DESC LIMIT 10;

SELECT GROUP_CONCAT(DISTINCT media.id ORDER BY media.id ASC SEPARATOR ',') AS media_ids,
       GROUP_CONCAT(DISTINCT media.path ORDER BY media.id ASC SEPARATOR ',') AS media_paths,
       GROUP_CONCAT(DISTINCT media.type ORDER BY media.id ASC SEPARATOR ',') AS media_types,
       GROUP_CONCAT(DISTINCT media.status ORDER BY media.id ASC SEPARATOR ',') AS media_statuses FROM media LEFT JOIN posts ON media.post_id = posts.id WHERE post_id = 6;



SELECT
    posts.id, posts.content, posts.created_at, posts.updated_at,
    GROUP_CONCAT(DISTINCT categories.id ORDER BY categories.id ASC SEPARATOR ',') AS category_ids,
    GROUP_CONCAT(DISTINCT categories.name ORDER BY categories.id ASC SEPARATOR ',') AS category_names,
    GROUP_CONCAT(DISTINCT media.id ORDER BY media.id ASC SEPARATOR ',') AS media_ids,
    GROUP_CONCAT(DISTINCT media.path ORDER BY media.id ASC SEPARATOR ',') AS media_paths,
    GROUP_CONCAT(DISTINCT media.type ORDER BY media.id ASC SEPARATOR ',') AS media_types,
    GROUP_CONCAT(DISTINCT media.status ORDER BY media.id ASC SEPARATOR ',') AS media_statuses FROM `posts`
    LEFT JOIN post_categories ON post_categories.post_id = posts.id
    LEFT JOIN categories ON categories.id = post_categories.category_id
    LEFT JOIN media ON media.post_id = posts.id WHERE posts.id = '6' GROUP BY `posts`.`id`

