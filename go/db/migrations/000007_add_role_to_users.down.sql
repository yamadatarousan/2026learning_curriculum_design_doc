-- usersテーブルからroleカラムを削除します
ALTER TABLE users DROP COLUMN IF EXISTS role;
