-- usersテーブルにroleカラムを追加します。デフォルトは'user'とします。
ALTER TABLE users ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'user';
