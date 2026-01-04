-- パスワードはすべて 'password123'
-- bcrypt hash for 'password123': $2a$10$5K/yM.o1V..c5P5WHC2v5.k.b4D5Y2s3s4b5E6f7G8h9i0j1k2l3m

-- テナントはまだ導入していないので、ユーザーとTODOのみ作成

-- Admin User (ID: 1)
INSERT INTO users (id, email, password_hash, role) 
VALUES (1, 'admin-test@example.com', '$2a$10$5K/yM.o1V..c5P5WHC2v5.k.b4D5Y2s3s4b5E6f7G8h9i0j1k2l3m', 'admin')
ON CONFLICT (id) DO NOTHING;

-- Normal User (ID: 2)
INSERT INTO users (id, email, password_hash, role) 
VALUES (2, 'user-test@example.com', '$2a$10$5K/yM.o1V..c5P5WHC2v5.k.b4D5Y2s3s4b5E6f7G8h9i0j1k2l3m', 'user')
ON CONFLICT (id) DO NOTHING;

-- Normal User's Todo
INSERT INTO todos (name, user_id)
VALUES ('Todo for user 2', 2);

-- IDのシーケンスがずれないように、手動挿入したIDの最大値に更新する
SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));
