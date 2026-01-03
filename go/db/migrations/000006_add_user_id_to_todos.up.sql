-- todosテーブルにuser_idカラムを追加し、usersテーブルへの外部キー制約を設定します
ALTER TABLE todos ADD COLUMN user_id INTEGER;

ALTER TABLE todos 
ADD CONSTRAINT fk_user
FOREIGN KEY (user_id) 
REFERENCES users(id)
ON DELETE CASCADE; -- ユーザーが削除されたら、そのユーザーのTODOも一緒に削除する
