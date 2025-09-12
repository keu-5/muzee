# Muzee Backend

Go + PostgreSQL + ent を使用した型安全なバックエンド API

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

# ent (型安全 ORM)
go get entgo.io/ent/cmd/ent
```

## 技術スタック

- **Go 1.25.1**
- **PostgreSQL** - データベース
- **pgx/v5** - PostgreSQL ドライバー
- **ent** - 型安全な ORM
- **goose** - データベースマイグレーション
- **Fiber v2** - Web フレームワーク
- **JWT** - 認証
- **Wire** - 依存性注入（DI）
- **zap** - 構造化ロギング

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
    NewClient,
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

このプロジェクトでは、ent ORM と Repository パターンを使用してデータアクセス層を構築しています。

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

#### 2. ent スキーマ定義

`ent/schema/user.go` にスキーマを定義：

```go
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"time"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").Positive(),
		field.String("username").Unique().NotEmpty(),
		field.String("email").Unique().NotEmpty(),
		field.String("password").NotEmpty().Sensitive(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("deleted_at"),
	}
}
```

#### 3. ent コード生成

```bash
# 型安全なGoコードを生成
go generate ./ent
```

#### 4. Repository 層の実装

```go
package repository

import (
	"context"
	"time"

	"github.com/keu-5/muzee/backend/ent"
	"github.com/keu-5/muzee/backend/ent/user"
)

type UserRepository struct {
	client *ent.Client
}

func NewUserRepository(client *ent.Client) *UserRepository {
	return &UserRepository{
		client: client,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, username, email, password string) (*ent.User, error) {
	return r.client.User.Create().
		SetUsername(username).
		SetEmail(email).
		SetPassword(password).
		Save(ctx)
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int) (*ent.User, error) {
	return r.client.User.Query().
		Where(user.ID(id), user.DeletedAtIsNil()).
		Only(ctx)
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*ent.User, error) {
	return r.client.User.Query().
		Where(user.Username(username), user.DeletedAtIsNil()).
		Only(ctx)
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*ent.User, error) {
	return r.client.User.Query().
		Where(user.Email(email), user.DeletedAtIsNil()).
		Only(ctx)
}

func (r *UserRepository) UpdateUser(ctx context.Context, id int, username, email, password string) (*ent.User, error) {
	return r.client.User.UpdateOneID(id).
		SetUsername(username).
		SetEmail(email).
		SetPassword(password).
		Save(ctx)
}

func (r *UserRepository) DeleteUser(ctx context.Context, id int) error {
	return r.client.User.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
```

#### 5. Service 層での使用

```go
package service

import (
	"context"
	"errors"

	"github.com/keu-5/muzee/backend/ent"
	"github.com/keu-5/muzee/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repository.UserRepository
	config   *config.Config
}

func NewAuthService(userRepo *repository.UserRepository, config *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		config:   config,
	}
}

func (s *AuthService) CreateUser(ctx context.Context, username, email, password string) (*ent.User, error) {
	// ユーザー名の重複チェック
	existingUser, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// パスワードハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return s.userRepo.CreateUser(ctx, username, email, string(hashedPassword))
}
```

## 設定

### ent 設定

ent では設定ファイルは不要で、Go のコードでスキーマを定義します。

```go
// ent/generate.go
//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate ./schema
package ent
```

### データベース接続

```go
package database

import (
	"database/sql"
	"fmt"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/keu-5/muzee/backend/config"
	"github.com/keu-5/muzee/backend/ent"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var EntClient *ent.Client

func ConnectDatabase(cfg *config.Config) (*ent.Client, error) {
	var dsn string

	if cfg.DatabaseURL != "" {
		dsn = cfg.DatabaseURL
	} else {
		dsn = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.DatabaseUser, cfg.DatabasePass, cfg.DatabaseHost, cfg.DatabasePort, cfg.DatabaseName)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(30)
	db.SetMaxIdleConns(5)

	drv := entsql.OpenDB("postgres", db)
	client := ent.NewClient(ent.Driver(drv))

	EntClient = client
	return client, nil
}
```

## Repository 層の設計原則

### 1. コンテキスト伝播

- 全てのメソッドで `context.Context` を第一引数として受け取る
- タイムアウトやキャンセレーション制御が可能

### 2. エラーハンドリング

- `ent.IsNotFound(err)` を使用してレコード未発見をハンドリング
- レコードが見つからない場合は nil を返す

### 3. 型安全性

- ent により生成された struct とメソッドを使用
- コンパイル時に型チェックが行われる
- クエリビルダーで型安全なクエリ構築

### 4. テスタビリティ

- インターフェースを定義してモックテストが可能
- 依存注入パターンを採用

### 5. パフォーマンス

- 接続プールを使用
- Eager Loading で N+1 問題を回避
- インデックス設定をスキーマで定義

## ロギング

このプロジェクトでは、Uber の zap パッケージを使用して構造化ロギングを実装しています。

### ロガーの初期化

```go
// サーバー起動時にロガーを初期化
logger.Init()
```

### ログレベル

以下のログレベルが利用可能です：

- **Debug** - デバッグ情報
- **Info** - 一般的な情報
- **Warn** - 警告
- **Error** - エラー
- **Fatal** - 致命的なエラー（ログ出力後にプログラムが終了）

### 使用例

```go
import (
    "github.com/keu-5/muzee/backend/pkg/logger"
    "go.uber.org/zap"
)

// 基本的なログ出力
logger.Info("Server starting")

// フィールド付きのログ出力
logger.Info("User login",
    zap.String("username", user.Name),
    zap.Int("user_id", user.ID),
)

// エラー情報を含むログ
if err != nil {
    logger.Error("Failed to process request", zap.Error(err))
}
```

### ログ出力形式

ログは JSON 形式で出力され、以下のフィールドが含まれます：

```json
{
  "level": "info",
  "time": "2025-09-13T00:00:00.000Z",
  "caller": "app/main.go:42",
  "msg": "Server starting",
  "user_id": 123,
  "username": "john_doe"
}
```

この形式により、ログの収集、検索、分析が容易になります。

## 開発コマンド

```bash
# マイグレーションファイル作成
goose -dir sql/schema create table_name sql

# ent コード生成
go generate ./ent

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

この設計により、型安全で保守性の高いデータアクセス層を ent ORM で構築できます。
