-- todosテーブルのnameカラムにUNIQUE制約を追加します
ALTER TABLE todos ADD CONSTRAINT todos_name_unique UNIQUE (name);
