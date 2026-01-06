# Runbook: TODO API 障害対応

## 概要

このRunbookは、TODO APIで障害が発生した際の初動対応手順を記載しています。
障害発生時は本書に従って対応し、復旧後は必ずポストモーテムを実施してください。

## 連絡先・エスカレーション

| 役割 | 担当者 | 連絡先 | 稼働時間 |
|------|--------|--------|----------|
| オンコール担当 | - | - | 24/7 |
| バックアップ担当 | - | - | 24/7 |
| インフラ担当 | - | - | 平日9-18時 |
| プロダクトオーナー | - | - | 平日9-18時 |

### エスカレーション基準

- **即座にエスカレーション**: サービス全体停止、データ損失の恐れ
- **30分以内にエスカレーション**: 復旧の見込みが立たない場合
- **1時間以内にエスカレーション**: 部分的障害が継続する場合

## 事前準備（チェックリスト）

対応前に以下を確認してください：

- [ ] ログへのアクセス権限
- [ ] DBへの読み取り権限（本番は慎重に）
- [ ] サーバーへのSSH接続
- [ ] モニタリングダッシュボードへのアクセス
- [ ] デプロイ履歴の確認方法
- [ ] ロールバック手順の把握

---

## 障害パターン別対応

### 1. サーバーが起動しない / サービスが落ちている

#### 症状
- `/health` エンドポイントが応答しない
- 502 Bad Gateway / 503 Service Unavailable
- プロセスが起動していない

#### 初動確認（5分以内）

```bash
# 1. プロセス確認
ps aux | grep app

# 2. コンテナ確認（Docker環境の場合）
docker ps -a
docker logs <container_id> --tail 100

# 3. システムログ確認
sudo journalctl -u todo-api -n 100 --no-pager

# 4. ディスク容量確認
df -h

# 5. メモリ確認
free -h
```

#### 原因と対応

| 原因 | 確認方法 | 対応 |
|------|----------|------|
| **設定ミス** | ログに "config error" や "failed to load" | 設定ファイルを確認・修正、再起動 |
| **DB接続失敗** | ログに "failed to connect to database" | DB状態確認 → セクション3参照 |
| **ポート競合** | ログに "address already in use" | 既存プロセスを停止、または別ポートで起動 |
| **メモリ不足** | OOMKillerログ、dmesg | メモリを増やす、または他プロセスを停止 |
| **ディスク容量** | df -h で使用率100% | 不要なログ/ファイルを削除 |

#### 復旧手順

```bash
# アプリケーション再起動
sudo systemctl restart todo-api

# Docker環境の場合
docker-compose restart app

# 起動確認
curl http://localhost:8080/health

# ログ監視
tail -f /var/log/todo-api/app.log
# または
docker logs -f <container_id>
```

#### 復旧しない場合

1. 直前のデプロイをロールバック（セクション5参照）
2. バックアップ担当にエスカレーション

---

### 2. サービスが遅い / タイムアウトが発生している

#### 症状
- リクエストが30秒以上かかる
- タイムアウトエラーが頻発
- レスポンスタイムが通常の10倍以上

#### 初動確認（5分以内）

```bash
# 1. CPU/メモリ使用率確認
top
htop  # 利用可能な場合

# 2. 接続数確認
netstat -an | grep :8080 | wc -l

# 3. DB接続確認
# PostgreSQL
psql -h localhost -U user -d todo_db -c "SELECT count(*) FROM pg_stat_activity;"

# 4. スロークエリログ確認
# PostgreSQL
psql -h localhost -U user -d todo_db -c "SELECT pid, now() - pg_stat_activity.query_start AS duration, query FROM pg_stat_activity WHERE state = 'active' ORDER BY duration DESC LIMIT 5;"

# 5. アプリケーションログ確認
tail -n 100 /var/log/todo-api/app.log | grep "ERROR\|WARN"
```

#### 原因と対応

| 原因 | 確認方法 | 対応 |
|------|----------|------|
| **DB接続枯渇** | ログに "connection pool exhausted" | DB接続プールを増やす、またはアプリ再起動 |
| **スロークエリ** | スロークエリログに長時間クエリ | クエリを特定 → EXPLAIN実行 → インデックス追加検討 |
| **CPU/メモリ高負荷** | top/htopで90%超え | スケールアップ or スケールアウト検討 |
| **外部API遅延** | ログにタイムアウトエラー | 外部サービス状態確認、リトライ設定確認 |
| **大量リクエスト** | アクセスログで急激な増加 | レートリミット確認、DDoS可能性を調査 |

#### 応急処置

```bash
# 1. 接続プール設定を緩和（一時的）
# 環境変数で設定している場合
export DB_MAX_CONNECTIONS=50

# 2. アプリケーション再起動
sudo systemctl restart todo-api

# 3. 負荷軽減（必要に応じて）
# - CDNのキャッシュTTLを延長
# - 非本質的な機能を一時的に無効化
```

#### 恒久対応

- スロークエリの最適化
- インデックス追加
- キャッシュ導入
- 水平スケーリング

---

### 3. DBに繋がらない

#### 症状
- ログに "connection refused" または "timeout"
- API が 500 Internal Server Error を返す
- `/health` は成功するが、TODO操作が全て失敗

#### 初動確認（5分以内）

```bash
# 1. DBプロセス確認
ps aux | grep postgres

# Docker環境の場合
docker ps -a | grep postgres

# 2. DB接続テスト
psql -h localhost -U user -d todo_db -c "SELECT 1;"

# 3. DBログ確認
# PostgreSQL
sudo tail -n 100 /var/log/postgresql/postgresql-*.log

# Docker環境の場合
docker logs <postgres_container_id> --tail 100

# 4. ネットワーク確認
ping <db_host>
telnet <db_host> 5432

# 5. DB接続数確認
psql -h localhost -U user -d todo_db -c "SELECT count(*) FROM pg_stat_activity;"
psql -h localhost -U user -d todo_db -c "SELECT max_connections FROM pg_settings WHERE name='max_connections';"
```

#### 原因と対応

| 原因 | 確認方法 | 対応 |
|------|----------|------|
| **DB停止** | psコマンドでプロセスなし | DB再起動 |
| **接続数上限** | 接続数がmax_connectionsに達している | 既存接続をkillまたはmax_connections増加 |
| **認証エラー** | ログに "authentication failed" | 認証情報を確認、修正 |
| **ネットワーク問題** | pingやtelnetが失敗 | ファイアウォール、セキュリティグループ確認 |
| **ディスク容量** | df -h で100% | ログやデータを削除 |

#### 復旧手順

```bash
# DB再起動
sudo systemctl restart postgresql

# Docker環境の場合
docker-compose restart db

# 接続確認
psql -h localhost -U user -d todo_db -c "SELECT 1;"

# アプリケーション再起動（接続プールをリセット）
sudo systemctl restart todo-api
```

#### 接続数が上限に達している場合

```sql
-- 現在の接続を確認
SELECT pid, usename, application_name, client_addr, state, query
FROM pg_stat_activity
WHERE datname = 'todo_db';

-- idle状態の接続をkill（慎重に）
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = 'todo_db' AND state = 'idle' AND state_change < now() - interval '10 minutes';
```

---

### 4. 特定のエンドポイントだけが失敗する

#### 症状
- `/todos` だけが500エラー
- 他のエンドポイントは正常

#### 初動確認（5分以内）

```bash
# 1. エラーログを絞り込み
grep "POST /api/v1/todos" /var/log/todo-api/app.log | tail -n 20

# 2. 該当エンドポイントをテスト
curl -X POST http://localhost:8080/api/v1/todos \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"test"}'

# 3. DB状態確認（該当テーブル）
psql -h localhost -U user -d todo_db -c "SELECT count(*) FROM todos;"
psql -h localhost -U user -d todo_db -c "\d todos"
```

#### 原因と対応

| 原因 | 確認方法 | 対応 |
|------|----------|------|
| **マイグレーション未実行** | テーブル/カラムが存在しない | マイグレーション実行 |
| **データ不整合** | 外部キー制約違反 | データ修正またはロールバック |
| **バリデーションエラー** | ログに validation error | リクエスト内容を確認、修正 |
| **バグ** | ログにスタックトレース | バグ修正版をデプロイ or ロールバック |

---

### 5. デプロイ後に問題が発生

#### 即座にロールバック判断

以下の場合は調査よりもロールバック優先：
- エラー率が10%を超える
- 5分以内に原因が特定できない
- データ損失の恐れがある

#### ロールバック手順

```bash
# 1. 前バージョンのタグ確認
git log --oneline -10

# 2. ロールバック実行（例：Dockerタグ指定）
docker pull myapp:v1.2.3
docker-compose up -d

# または、Gitリポジトリから
git checkout v1.2.3
docker build -t myapp:latest .
docker-compose up -d

# 3. ヘルスチェック
curl http://localhost:8080/health

# 4. 動作確認
# 主要エンドポイントをテスト

# 5. ログ監視
docker logs -f <container_id>
```

---

## モニタリング・アラート

### 主要メトリクス

| メトリクス | 正常範囲 | 警告 | 緊急 |
|-----------|---------|------|------|
| レスポンスタイム（P95） | < 500ms | > 1s | > 3s |
| エラー率 | < 0.1% | > 1% | > 5% |
| CPU使用率 | < 60% | > 80% | > 95% |
| メモリ使用率 | < 70% | > 85% | > 95% |
| DB接続数 | < max * 0.7 | > max * 0.8 | > max * 0.9 |

### ログの場所

| ログ種別 | パス |
|---------|------|
| アプリケーションログ | `/var/log/todo-api/app.log` |
| アクセスログ | `/var/log/todo-api/access.log` |
| DBログ | `/var/log/postgresql/postgresql-*.log` |
| システムログ | `journalctl -u todo-api` |

---

## 復旧後の対応

### チェックリスト

- [ ] すべての主要機能が動作することを確認
- [ ] エラー率が正常範囲に戻ったことを確認
- [ ] 関係者に復旧を通知
- [ ] インシデントチケットを作成
- [ ] ポストモーテムをスケジュール（48時間以内）

### ポストモーテムで記録すべき内容

1. **発生日時**: いつ障害が発生したか
2. **検知方法**: どのように気づいたか（アラート、ユーザー報告など）
3. **影響範囲**: どのユーザー・機能が影響を受けたか
4. **根本原因**: なぜ発生したか（5 Whys）
5. **対応内容**: どう対処したか（時系列）
6. **改善策**: 再発防止のために何をするか（担当者・期限付き）

---

## よくある質問（FAQ）

### Q1. ロールバックすべきか判断に迷ったら？

**A:** 以下のいずれかに該当する場合は迷わずロールバック：
- エラー率が10%を超えている
- 5分以内に原因が特定できない
- ユーザーへの影響が大きい

迷ったら上長にエスカレーションしてください。

### Q2. 本番DBに直接接続してもいい？

**A:** 原則として**読み取り専用**で接続してください。
書き込みが必要な場合は必ず：
1. 別の担当者にレビューしてもらう
2. 変更内容をチケットに記録する
3. バックアップを取得する

### Q3. ログが見つからない

**A:** 以下を確認：
```bash
# Dockerの場合
docker logs <container_id>

# systemdの場合
journalctl -u todo-api

# ログローテーションされている可能性
ls -lh /var/log/todo-api/
zcat /var/log/todo-api/app.log.1.gz | tail -n 100
```

---

## 付録：便利なコマンド集

### システム情報

```bash
# ディスク使用状況
df -h
du -sh /var/log/* | sort -h

# メモリ使用状況
free -h
ps aux --sort=-%mem | head -10

# CPU使用状況
top -bn1 | head -20
ps aux --sort=-%cpu | head -10

# ネットワーク接続
netstat -tuln | grep LISTEN
ss -tuln | grep :8080
```

### Docker

```bash
# コンテナ一覧
docker ps -a

# ログ確認
docker logs <container_id> -f --tail 100

# コンテナ内に入る
docker exec -it <container_id> /bin/sh

# リソース使用状況
docker stats
```

### PostgreSQL

```bash
# 接続テスト
psql -h localhost -U user -d todo_db -c "SELECT version();"

# テーブル一覧
psql -h localhost -U user -d todo_db -c "\dt"

# アクティブなクエリ
psql -h localhost -U user -d todo_db -c "SELECT pid, now() - pg_stat_activity.query_start AS duration, query, state FROM pg_stat_activity WHERE state != 'idle' ORDER BY duration DESC;"

# 接続数
psql -h localhost -U user -d todo_db -c "SELECT count(*) FROM pg_stat_activity;"
```

---

## 更新履歴

| 日付 | 更新者 | 内容 |
|------|--------|------|
| 2024-01-01 | - | 初版作成 |
