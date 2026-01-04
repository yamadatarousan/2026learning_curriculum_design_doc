-- パスワードはすべて 'password123'
-- bcrypt hash for 'password123': $2a$10$kxtxAB6YnV5vub0dbnc9z.DmL92hzshSp/X32LFR8G8//BxSx2Us6

-- テナントはまだ導入していないので、ユーザーとTODOのみ作成

-- Admin User (ID: 1)
INSERT INTO users (id, email, password_hash, role) 
VALUES (1, 'admin-test@example.com', '$2a$10$kxtxAB6YnV5vub0dbnc9z.DmL92hzshSp/X32LFR8G8//BxSx2Us6', 'admin')
ON CONFLICT (id) DO NOTHING;

-- Normal User (ID: 2)
INSERT INTO users (id, email, password_hash, role) 
VALUES (2, 'user-test@example.com', '$2a$10$kxtxAB6YnV5vub0dbnc9z.DmL92hzshSp/X32LFR8G8//BxSx2Us6', 'user')
ON CONFLICT (id) DO NOTHING;

-- Normal User's Todo
INSERT INTO todos (name, user_id)
VALUES ('Todo for user 2', 2);

-- IDのシーケンスがずれないように、手動挿入したIDの最大値に更新する
SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));
