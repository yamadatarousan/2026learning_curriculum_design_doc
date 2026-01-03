# Day 49: マルチテナント化 設計メモ

## 1. マルチテナントとは

一つのアプリケーションインスタンスを、複数の独立した顧客グループ（テナント）で共有するアーキテクチャ。各テナントのデータは他のテナントから完全に分離されている必要がある。

## 2. 設計方針

テナントの識別子として`tenant_id`を導入し、データベースレベルでデータの分離を強制する「スキーマ共有・データ分離」方式を採用する。

## 3. 設計変更案

### 3.1. データベースの変更

1.  **`tenants`テーブルの追加**
    各テナント（組織やワークスペース）の情報を管理するテーブルを新設する。
    ```sql
    CREATE TABLE tenants (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
    ```

2.  **`users`テーブルへの`tenant_id`追加**
    ユーザーがどのテナントに属しているかを管理する。最もシンプルなのは、1ユーザー=1テナントの想定で`users`テーブルに`tenant_id`を追加する方法。
    ```sql
    -- 1ユーザーが1テナントにのみ所属する場合
    ALTER TABLE users ADD COLUMN tenant_id INTEGER NOT NULL;
    ALTER TABLE users ADD CONSTRAINT fk_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id);
    ```
    > **発展:** 1ユーザーが複数テナントに所属できるようにするには、`user_tenants`という中間テーブル（`user_id`, `tenant_id`）を作成する方が拡張性が高い。

3.  **`todos`テーブルへの`tenant_id`追加**
    TODOデータもテナントごとに分離する。
    ```sql
    ALTER TABLE todos ADD COLUMN tenant_id INTEGER NOT NULL;
    ALTER TABLE todos ADD CONSTRAINT fk_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id);
    ```

### 3.2. 認証・認可 (JWT) の変更

1.  **JWTクレームへの`tenant_id`追加**
    ログイン成功時に発行するJWTのペイロードに、ユーザーが現在操作対象としている`tenant_id`を含める。
    ```go
    // AppClaims構造体を拡張
    type AppClaims struct {
        TenantID int    `json:"tid"` // Tenant ID
        Role     string `json:"role"`
        jwt.RegisteredClaims
    }
    ```

2.  **ログイン処理の変更**
    -   ログイン時にユーザー情報と共に所属テナントの情報を取得する。
    -   ユーザーが複数テナントに所属している場合は、どのテナントでログインするかを選択させるUI/APIが必要になる。
    -   選択されたテナントIDをJWTの`TenantID`クレームにセットしてトークンを生成する。

### 3.3. アプリケーションロジックの変更

1.  **`authMiddleware`の修正**
    -   JWTを検証する際に`tenant_id`もクレームから抽出し、Ginのコンテキストに保存する。
    -   `c.Set("tenantID", claims.TenantID)`

2.  **リポジトリ層の全面的な修正**
    -   **すべてのデータアクセス関数**（`FindAll`, `CreateTodoWithAudit`など）の引数に`tenantID int`を追加する。
    -   関数内の**すべてのSQLクエリ**に`WHERE tenant_id = $N`という条件句を追加し、必ずテナントでデータが絞り込まれるようにする。
    -   これを忘れると、他のテナントのデータが漏洩する重大なセキュリティインシデントに繋がる。

    **修正例 (`repository.go`):**
    ```go
    // 変更前
    func (r *TodoRepository) FindAll(userID int) ([]Todo, error) {
        rows, err := r.db.Query("SELECT ... FROM todos WHERE user_id = $1", userID)
        // ...
    }

    // 変更後
    func (r *TodoRepository) FindAll(tenantID int, userID int) ([]Todo, error) {
        rows, err := r.db.Query("SELECT ... FROM todos WHERE tenant_id = $1 AND user_id = $2", tenantID, userID)
        // ...
    }
    ```

## 4. まとめ

マルチテナント化は、単純に`tenant_id`カラムを追加するだけでなく、認証、認可、データアクセスの全てのレイヤーに影響を及ぼす、アプリケーション全体に関わる大きな設計変更である。特に、リポジトリ層でのデータ絞り込みは徹底する必要がある。
