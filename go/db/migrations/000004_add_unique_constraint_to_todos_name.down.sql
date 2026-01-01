-- todosテーブルのnameカラムからUNIQUE制約を削除します
ALTER TABLE todos DROP CONSTRAINT IF EXISTS todos_name_unique;
