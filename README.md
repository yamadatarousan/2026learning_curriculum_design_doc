# Go TODO API Server

これは、30日間のGo言語学習カリキュラムの成果物として作成された、シンプルなTODOリストAPIサーバーです。

このアプリケーションは、Goの標準ライブラリといくつかの一般的な外部ライブラリを使用して、完全なCRUD（作成、読み取り、更新、削除）機能を提供します。データはSQLiteデータベースに永続的に保存されます。

## 機能

- TODOアイテムの作成 (Create)
- TODOアイテムのリスト取得 (Read)
- TODOアイテムの更新 (Update)
- TODOアイテムの削除 (Delete)

## 要件

- Go (バージョン 1.18 以上を推奨)

## 実行方法

1.  **依存関係のインストール:**
    プロジェクトのルートディレクトリで以下のコマンドを実行し、必要なライブラリをインストールします。
    ```bash
    go mod tidy
    ```

2.  **サーバーの起動:**
    `day30` ディレクトリ（あるいは `main.go` があるディレクトリ）に移動し、以下のコマンドを実行します。
    ```bash
    go run .
    ```
    サーバーが `http://localhost:8080` で起動します。

## API仕様

### 1. TODOリストの取得

- **エンドポイント:** `GET /todos`
- **説明:** すべてのTODOアイテムのリストをJSON形式で返します。
- **成功レスポンス (200 OK):**
  ```json
  [
    {
      "id": 1,
      "name": "Learn Go"
    },
    {
      "id": 2,
      "name": "Write a README"
    }
  ]
  ```

### 2. TODOの新規作成

- **エンドポイント:** `POST /todos`
- **説明:** 新しいTODOアイテムを作成します。
- **リクエストボディ:**
  ```json
  {
    "name": "My New Todo"
  }
  ```
- **成功レスポンス (201 Created):**
  ```json
  {
    "id": 3,
    "name": "My New Todo"
  }
  ```

### 3. TODOの更新

- **エンドポイント:** `PUT /todos/{id}`
- **説明:** 指定したIDのTODOアイテムの名前を更新します。
- **リクエストボディ:**
  ```json
  {
    "name": "My Updated Todo Name"
  }
  ```
- **成功レスポンス (200 OK):**
  ```
  Todo updated successfully
  ```

### 4. TODOの削除

- **エンドポイント:** `DELETE /todos/{id}`
- **説明:** 指定したIDのTODOアイテムを削除します。
- **成功レスポンス (204 No Content):**
  (レスポンスボディなし)