# Day 50: Secrets（秘密情報）管理方針

本プロジェクトにおいて、Secrets（データベースのパスワード、APIキー、JWTの秘密鍵など）は以下のルールに従って管理する。

## 原則: ソースコードにハードコードしない

いかなる理由があっても、Secretsをソースコード（`.go`ファイルなど）に直接書き込むことを禁止する。GitリポジトリにSecretsが含まれる状況を絶対に避ける。

## 環境ごとの管理方法

### 1. 開発 (Development) 環境

-   **方法:** 環境変数（Environment Variables）を使用してSecretsをアプリケーションに渡す。
-   **実践:**
    1.  プロジェクトのルートに、Gitの管理対象外となる`.env`ファイルを作成する。
        ```.env
        # .env - このファイルは .gitignore に追加し、Git管理に含めない
        DB_PASSWORD=password
        JWT_SECRET=a-very-secret-key-for-development
        ```
    2.  Goのライブラリ（例: `github.com/joho/godotenv`）を使い、アプリケーション起動時に`.env`ファイルの内容を環境変数として読み込む。
    3.  Gitリポジトリには、どのような環境変数が必要かを示すためのサンプルファイル (`.env.example`) を含める。
        ```.env.example
        # .env.example - このファイルはGit管理に含める
        DB_PASSWORD=
        JWT_SECRET=
        ```

### 2. 本番 (Production) 環境

-   **方法:** **シークレット管理サービス** の利用を第一選択とする。
-   **実践:**
    -   **推奨:** Google Secret Manager, AWS Secrets Manager, HashiCorp Vault などの専用サービスにSecretsを保管する。アプリケーションは、適切なIAMロールや認証情報を元に、起動時や必要時にこれらのサービスから直接Secretsを取得する。
    -   **次善策:** コンテナの実行環境（Kubernetes Secrets, Docker Composeのenvironmentセクションなど）から環境変数として注入する。この場合、環境変数の値そのものを誰が閲覧できるか、厳格に管理する必要がある。

## まとめ

-   **コードと設定（特にSecrets）は分離する。**
-   **開発環境では`.env`ファイルを活用する。**
-   **本番環境では専用のシークレット管理サービスの利用を強く推奨する。**
