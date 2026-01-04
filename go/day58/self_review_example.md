# セルフレビュー実施例

このドキュメントは、Day 54で作成したコード（Gin + PostgreSQL + JWT認証のTODO API）に対してセルフレビューを行った結果の例です。

---

## レビュー対象

- ファイル: `go/day54/main.go`
- 内容: JWT認証付きTODO API
- レビュー日: 2026-01-04

---

## 発見された問題点と改善提案

### 🔴 Critical（致命的）

#### 1. JWTシークレットがハードコードされている

**問題箇所**:
```go
var jwtSecret = []byte("a-very-secret-key")  // line 42
```

**問題点**:
- 機密情報がコードに直接書かれている
- Gitにコミットされると第三者に漏れる
- シークレットの変更が困難

**改善案**:
```go
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// または、起動時にチェック
func init() {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        log.Fatal("JWT_SECRET environment variable is required")
    }
    jwtSecret = []byte(secret)
}
```

**優先度**: 🔴 最優先（本番環境では必須）

---

#### 2. エラーの詳細を外部に漏らしている

**問題箇所**:
```go
c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
    "error": "Invalid token",
    "details": err.Error()  // 内部エラーの詳細を返している
})
```

**問題点**:
- エラーの詳細が攻撃者の手がかりになる
- 内部実装の情報が漏れる可能性

**改善案**:
```go
// ログには詳細を出力
log.Printf("Token validation failed: %v", err)

// レスポンスには最小限の情報のみ
c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
    "error": "Invalid token"
})
```

**優先度**: 🔴 高（セキュリティリスク）

---

### 🟡 High（高）

#### 3. グローバル変数 `db` の使用

**問題箇所**:
```go
var db *sql.DB  // line 41
```

**問題点**:
- テストが困難（モックに差し替えられない）
- 依存関係が不明確

**改善案**:
```go
// Repository構造体にDB接続を持たせる
type TodoRepository struct {
    db *sql.DB
}

func NewTodoRepository(db *sql.DB) *TodoRepository {
    return &TodoRepository{db: db}
}

// Handlerに注入
type TodoHandler struct {
    repo *TodoRepository
}

func NewTodoHandler(repo *TodoRepository) *TodoHandler {
    return &TodoHandler{repo: repo}
}
```

**優先度**: 🟡 高（設計改善）

---

#### 4. ハンドラが太すぎる（責務が多い）

**問題点**:
- ハンドラにビジネスロジックとDBアクセスが混在
- テストが困難
- 再利用性が低い

**改善案**:
レイヤーを分離する
```
Handler → Service → Repository
```

```go
// Service層を追加
type TodoService struct {
    repo *TodoRepository
}

func (s *TodoService) CreateTodo(name string, userID int) (*Todo, error) {
    // ビジネスロジック（バリデーション、変換など）
    if name == "" {
        return nil, errors.New("name is required")
    }

    return s.repo.Create(name, userID)
}

// Handlerは薄く
func (h *TodoHandler) createTodo(c *gin.Context) {
    var req TodoRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    userID := getUserIDFromContext(c)
    todo, err := h.service.CreateTodo(req.Name, userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, todo)
}
```

**優先度**: 🟡 高（設計改善）

---

### 🟢 Medium（中）

#### 5. エラーレスポンスの形式が統一されていない

**問題箇所**:
```go
// 場所によってエラーレスポンスの形式が異なる
gin.H{"error": "message"}
gin.H{"error": "message", "details": "..."}
gin.H{"error": "message", "message": "..."}
```

**改善案**:
```go
// エラーレスポンスの構造体を定義
type ErrorResponse struct {
    Error   string `json:"error"`
    Code    string `json:"code,omitempty"`
    Message string `json:"message,omitempty"`
}

// ヘルパー関数
func respondError(c *gin.Context, status int, code, message string) {
    c.JSON(status, ErrorResponse{
        Error:   message,
        Code:    code,
    })
}

// 使用例
respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token")
```

**優先度**: 🟢 中（一貫性向上）

---

#### 6. バリデーションエラーのメッセージが不親切

**問題点**:
```go
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 出力: "Key: 'Todo.Name' Error:Field validation for 'Name' failed on the 'required' tag"
```

**改善案**:
```go
func formatValidationError(err error) map[string]string {
    errors := make(map[string]string)

    if validationErrors, ok := err.(validator.ValidationErrors); ok {
        for _, e := range validationErrors {
            field := strings.ToLower(e.Field())
            switch e.Tag() {
            case "required":
                errors[field] = fmt.Sprintf("%s is required", field)
            case "email":
                errors[field] = fmt.Sprintf("%s must be a valid email", field)
            // ...
            default:
                errors[field] = fmt.Sprintf("%s is invalid", field)
            }
        }
    }

    return errors
}

// 使用例
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{
        "error": "Validation failed",
        "fields": formatValidationError(err),
    })
    return
}
```

**優先度**: 🟢 中（ユーザー体験向上）

---

#### 7. ログが不足している

**問題点**:
- 重要な処理（ユーザー登録、ログインなど）にログがない
- デバッグやトラブルシューティングが困難

**改善案**:
```go
func (h *AuthHandler) signup(c *gin.Context) {
    var req SignupRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        log.Printf("Signup validation failed: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    log.Printf("New user signup attempt: email=%s", req.Email)

    // ...

    log.Printf("User created successfully: id=%d, email=%s", user.ID, user.Email)
    c.JSON(http.StatusCreated, user)
}
```

**優先度**: 🟢 中（運用性向上）

---

### ⚪ Low（低）

#### 8. マジックナンバー

**問題箇所**:
```go
time.Now().Add(24 * time.Hour)  // JWTの有効期限
```

**改善案**:
```go
const (
    JWTExpirationHours = 24
)

expiresAt := time.Now().Add(JWTExpirationHours * time.Hour)
```

**優先度**: ⚪ 低（可読性向上）

---

#### 9. コメントが少ない

**改善案**:
```go
// authMiddleware はJWTトークンを検証し、有効な場合はclaimsをcontextに保存する
func authMiddleware() gin.HandlerFunc {
    // ...
}

// adminMiddleware はユーザーがadminロールを持つことを確認する
// 注意: authMiddlewareの後に配置すること
func adminMiddleware() gin.HandlerFunc {
    // ...
}
```

**優先度**: ⚪ 低（可読性向上）

---

## まとめ

### 発見された問題の数

| 優先度 | 件数 |
|--------|------|
| 🔴 Critical | 2件 |
| 🟡 High | 2件 |
| 🟢 Medium | 3件 |
| ⚪ Low | 2件 |
| **合計** | **9件** |

### 優先的に修正すべき項目（Top 3）

1. **JWTシークレットを環境変数から読み込む**（セキュリティ）
2. **エラーの詳細を外部に漏らさない**（セキュリティ）
3. **レイヤー分離（Handler/Service/Repository）**（設計）

### 次のアクション

- [ ] JWTシークレットを環境変数化
- [ ] エラーハンドリングを改善
- [ ] レイヤー分離の設計を検討
- [ ] エラーレスポンスの形式を統一
- [ ] ログを追加

---

## セルフレビューから学んだこと

### 1. セキュリティは最優先

- 機密情報のハードコード
- エラー情報の漏洩
- これらは本番環境では致命的

### 2. 設計の重要性

- グローバル変数は避ける
- レイヤーを分離する
- テストしやすい設計

### 3. ユーザー体験

- エラーメッセージは親切に
- レスポンスの一貫性

### 4. 運用性

- ログは重要
- デバッグしやすい設計

---

**レビュー完了日**: 2026-01-04
