# Muzee Backend

Go + PostgreSQL + sqlc を使用した型安全なバックエンド API

## セットアップ

### 1. Go 環境設定

`.zshrc` (または `.bashrc`) に GOPATH 設定を追加：

```bash
# Go
export GOPATH=$(go env GOPATH)
export PATH="$GOPATH/bin:$PATH"
```

設定を反映：

```bash
source ~/.zshrc  # または source ~/.bashrc
```

### 2. 必要なツールのインストール

```bash
# goose (マイグレーション)
go install github.com/pressly/goose/v3/cmd/goose@latest

# sqlc (型安全コード生成)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

## 技術スタック

- **Go 1.25.1**
- **PostgreSQL** - データベース
- **pgx/v5** - PostgreSQL ドライバー（接続プール付き）
- **sqlc** - 型安全な SQL コード生成
- **goose** - データベースマイグレーション
- **Fiber v2** - Web フレームワーク
- **JWT** - 認証
- **Wire** - 依存性注入（DI）

## アーキテクチャ

このプロジェクトは、クリーンアーキテクチャと Google Wire による依存性注入を採用しています。

### レイヤー構成

```mermaid
graph LR
    Handler --> Service --> Repository --> Database
```

各レイヤーは以下の責務を持ちます：

- **Handler**: HTTP リクエストの受付とレスポンスの返却
- **Service**: ビジネスロジックの実装
- **Repository**: データアクセスの抽象化
- **Database**: 実際のデータベース操作

### 依存性注入（DI）の実装

Google Wire を使用して、各レイヤー間の依存関係を自動的に解決します。

#### プロバイダーの設定

各レイヤーごとに Wire プロバイダーを定義：

```go
// Database (internal/database/wire.go)
var ProviderSet = wire.NewSet(
    ConnectDatabase,
    NewQueries,
)

// Repository (internal/repository/wire.go)
var ProviderSet = wire.NewSet(
    NewUserRepository,
)

// Service (internal/service/wire.go)
var ProviderSet = wire.NewSet(
    NewAuthService,
)

// Handler (internal/handler/wire.go)
var ProviderSet = wire.NewSet(
    NewHandler,
)
```

#### 依存関係の定義

`wire.go`で依存関係を定義し、`wire`コマンドで初期化コードを自動生成：

```go
//+build wireinject

func InitializeApp(cfg *config.Config) (*handler.Handler, error) {
    wire.Build(
        database.ProviderSet,
        repository.ProviderSet,
        service.ProviderSet,
        handler.ProviderSet,
    )
    return nil, nil
}
```

### テスタビリティ

DI パターンの採用により、各レイヤーは独立してテスト可能です：

- **Handler**: Service をモック
- **Service**: Repository をモック
- **Repository**: Database をモック

## データベース設計パターン

### Repository 層の作り方

このプロジェクトでは、Repository パターンを使用してデータアクセス層を構築しています。

#### 1. マイグレーションファイルの作成

```bash
# sql/schema ディレクトリにgooseマイグレーションファイルを作成
goose -dir sql/schema create add_users_table sql
```

作成されたファイルにテーブル定義を記述：

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_users_deleted_at ON users(deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_deleted_at;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
```

#### 2. SQL クエリの定義

`sql/queries/` ディレクトリにクエリファイルを作成：

```sql
-- name: CreateUser :one
INSERT INTO users (username, email, password, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING id, username, email, password, created_at, updated_at, deleted_at;

-- name: GetUserByID :one
SELECT id, username, email, password, created_at, updated_at, deleted_at
FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByUsername :one
SELECT id, username, email, password, created_at, updated_at, deleted_at
FROM users
WHERE username = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT id, username, email, password, created_at, updated_at, deleted_at
FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: UpdateUser :one
UPDATE users
SET username = $2, email = $3, password = $4, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, username, email, password, created_at, updated_at, deleted_at;

-- name: DeleteUser :exec
UPDATE users
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;
```

#### 3. sqlc コード生成

```bash
# 型安全なGoコードを生成
sqlc generate
```

#### 4. Repository 層の実装

```go
package repository

import (
    "context"
    "errors"

    "github.com/jackc/pgx/v5"
    "github.com/keu-5/muzee/backend/internal/database"
    "github.com/keu-5/muzee/backend/internal/db"
)

type UserRepository struct {
    queries *db.Queries
}

func NewUserRepository() *UserRepository {
    return &UserRepository{
        queries: database.GetQueries(),
    }
}

// contextを受け取る設計
func (r *UserRepository) CreateUser(ctx context.Context, params db.CreateUserParams) (db.User, error) {
    return r.queries.CreateUser(ctx, params)
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int32) (db.User, error) {
    user, err := r.queries.GetUserByID(ctx, id)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return db.User{}, nil // レコードが見つからない場合
        }
        return db.User{}, err
    }
    return user, nil
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (db.User, error) {
    user, err := r.queries.GetUserByUsername(ctx, username)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return db.User{}, nil
        }
        return db.User{}, err
    }
    return user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
    user, err := r.queries.GetUserByEmail(ctx, email)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return db.User{}, nil
        }
        return db.User{}, err
    }
    return user, nil
}
```

#### 5. Service 層での使用

```go
package service

import (
    "context"
    "errors"

    "github.com/keu-5/muzee/backend/internal/db"
    "github.com/keu-5/muzee/backend/internal/repository"
    "golang.org/x/crypto/bcrypt"
)

type AuthService struct {
    userRepo *repository.UserRepository
    config   *config.Config
}

func NewAuthService(config *config.Config) *AuthService {
    return &AuthService{
        userRepo: repository.NewUserRepository(),
        config:   config,
    }
}

func (s *AuthService) CreateUser(ctx context.Context, username, email, password string) (db.User, error) {
    // ユーザー名の重複チェック
    existingUser, err := s.userRepo.GetUserByUsername(ctx, username)
    if err != nil {
        return db.User{}, err
    }
    if existingUser.ID != 0 {
        return db.User{}, errors.New("username already exists")
    }

    // パスワードハッシュ化
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return db.User{}, err
    }

    params := db.CreateUserParams{
        Username: username,
        Email:    email,
        Password: string(hashedPassword),
    }

    return s.userRepo.CreateUser(ctx, params)
}
```

## 設定

### sqlc.yaml

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "./sql/queries"
    schema: "./sql/schema"
    gen:
      go:
        package: "db"
        out: "./internal/db"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_pointers_for_null_types: true
```

### データベース接続（接続プール使用）

```go
package database

import (
    "context"
    "fmt"
    "log"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/keu-5/muzee/backend/config"
    "github.com/keu-5/muzee/backend/internal/db"
)

var Pool *pgxpool.Pool
var Queries *db.Queries

func ConnectDatabase(cfg *config.Config) (*pgxpool.Pool, *db.Queries) {
    var dsn string

    if cfg.DatabaseURL != "" {
        dsn = cfg.DatabaseURL
    } else {
        dsn = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
            cfg.DatabaseUser, cfg.DatabasePass, cfg.DatabaseHost, cfg.DatabasePort, cfg.DatabaseName)
    }

    poolConfig, err := pgxpool.ParseConfig(dsn)
    if err != nil {
        log.Fatal("Failed to parse database config:", err)
    }

    // 接続プール設定
    poolConfig.MaxConns = 30
    poolConfig.MinConns = 5

    pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    Pool = pool
    Queries = db.New(pool)
    return pool, Queries
}
```

## Repository 層の設計原則

### 1. コンテキスト伝播

- 全てのメソッドで `context.Context` を第一引数として受け取る
- タイムアウトやキャンセレーション制御が可能

### 2. エラーハンドリング

- `pgx.ErrNoRows` を適切にハンドリング
- レコードが見つからない場合は空の構造体を返す

### 3. 型安全性

- sqlc により生成された struct とメソッドを使用
- コンパイル時に型チェックが行われる

### 4. テスタビリティ

- インターフェースを定義してモックテストが可能
- 依存注入パターンを採用

### 5. パフォーマンス

- 接続プール（pgxpool）を使用
- N+1 問題を避けるクエリ設計

## 開発コマンド

```bash
# マイグレーションファイル作成
goose -dir sql/schema create table_name sql

# sqlc コード生成
sqlc generate

# 依存関係の整理
go mod tidy

# ビルド
go build -o server ./cmd/server/main.go

# 実行
./server

# テスト実行
go test ./...

# 開発サーバー起動（ホットリロード）
go run ./cmd/server/main.go
```

## マイグレーション実行

```bash
# データベース接続文字列例
export DATABASE_URL="postgresql://username:password@localhost:5432/dbname?sslmode=disable"

# マイグレーション実行
goose -dir migrations postgres "$DATABASE_URL" up

# ロールバック
goose -dir migrations postgres "$DATABASE_URL" down

# ステータス確認
goose -dir migrations postgres "$DATABASE_URL" status
```

この設計により、型安全で保守性の高いデータアクセス層を構築できます。
