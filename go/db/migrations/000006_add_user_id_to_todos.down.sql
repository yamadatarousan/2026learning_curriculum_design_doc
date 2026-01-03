-- todosテーブルから外部キー制約とuser_idカラムを削除します
ALTER TABLE todos DROP CONSTRAINT IF EXISTS fk_user;
ALTER TABLE todos DROP COLUMN IF EXISTS user_id;
