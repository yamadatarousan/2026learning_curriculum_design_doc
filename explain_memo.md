# Day 42: インデックスとEXPLAINの学習メモ

## 1. EXPLAINとは？

`EXPLAIN`は、SQLクエリがデータベース内部でどのように実行されるか（実行計画）を確認するためのコマンド。
パフォーマンスチューニングの第一歩として、「クエリが遅いとき、まずEXPLAINを見る」のが定石。

## 2. インデックス追加前の実行計画

`todos`テーブルの`name`カラムで検索したときの実行計画。

### 実行コマンド
```sql
EXPLAIN SELECT * FROM todos WHERE name = 'some_name';
```

### 実行結果
```
                        QUERY PLAN
----------------------------------------------------------
 Seq Scan on todos  (cost=0.00..1.04 rows=1 width=36)
   Filter: (name = 'some_name'::text)
```

- **`Seq Scan`**: Sequential Scan（シーケンシャルスキャン）の略。
- **意味**: テーブルの全レコードを先頭から順番にスキャンしている。データ量が増えると極端に遅くなる。

## 3. インデックス追加後の実行計画

`name`カラムにインデックスを追加（`CREATE INDEX idx_todos_name ON todos(name);`）した後の実行計画。

### 実行コマンド
```sql
EXPLAIN SELECT * FROM todos WHERE name = 'some_name';
```

### 実行結果
```
                                  QUERY PLAN
-------------------------------------------------------------------------------
 Index Scan using idx_todos_name on todos  (cost=0.00..8.27 rows=1 width=36)
   Index Cond: (name = 'some_name'::text)
```

- **`Index Scan`**: インデックスを利用してスキャンしたことを示す。
- **意味**: 全件スキャンを避け、索引を使って効率的にデータを探している。これにより、データ量が増えても高速な検索が維持される。

## まとめ

- インデックスは、特定のカラムでの検索パフォーマンスを劇的に向上させる。
- `EXPLAIN`を使うことで、クエリがインデックスを正しく利用できているかを確認できる。
- 検索条件（`WHERE`句）で頻繁に利用されるカラムには、インデックスの追加を検討するのが良い。
