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
UPDATE users SET role_id = 3 WHERE id = 2;
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
SELECT * FROM role_permissions;
SELECT * FROM access_tokens WHERE status = 1;