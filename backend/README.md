# Backend

## 概要

このディレクトリは **Go (Fiber)** を使用した Web API サーバーです。
開発時は **Air** を利用してホットリロードを実現しています。

## 構成

- **[Go](https://go.dev/)**
  シンプルで高速に動作するプログラミング言語。並行処理に強く、バックエンド API の開発に適しています。本プロジェクトでは Fiber や Air と組み合わせて利用します。

- **[Fiber](https://gofiber.io/)**
  高速でシンプルな Go の Web フレームワーク。ルーティングやミドルウェアを管理します。

- **[Air](https://github.com/air-verse/air)**
  開発中のソースコード変更を検知し、自動的に再ビルド & 再起動してくれるツール。

- **[ent](https://entgo.io/)**
  Go 製の ORM（Object Relational Mapper）。スキーマ定義から型安全なコードを自動生成し、DB 操作を効率化します。

- **[viper](https://github.com/spf13/viper)**
  設定管理ライブラリ。`.env` ファイルや環境変数から設定値を読み込み、開発・本番で環境を切り替えられるようにします。

- **[fx](https://github.com/uber-go/fx)**
  Uber 製の依存性注入 (DI) フレームワーク。アプリのライフサイクル管理やコンポーネントの組み立てを自動化します。

- **[zap](https://github.com/uber-go/zap)**
  高性能な構造化ロガー。開発・本番で異なるログ設定を使い分け、パフォーマンスと可読性を両立します。

- **[swaggo/swag](https://github.com/swaggo/swag)**
  Go のコードコメントから Swagger(OpenAPI v2) ドキュメントを自動生成するツール。
  Fiber ハンドラーにコメントを追加することで、API 仕様を常に最新状態に保つことができます。

- **[OpenAPI Generator](https://openapi-generator.tech/)**
  `openapitools/openapi-generator-cli:latest-release` の Docker イメージを利用し、Swagger から OpenAPI v3 仕様への変換や、各種言語のクライアント SDK の生成を行います。
  CI/CD パイプラインに組み込むことで、API 仕様とクライアントの同期を自動化できます。

## エンドポイント追加方法

このプロジェクトでは **ent (ORM)** と **Clean Architecture** をベースにエンドポイントを追加します。
以下に、新しいエンティティ（例: `Article`）を追加する手順を説明します。

---

### 1. ent のスキーマ作成

まずは ent のスキーマを作成します。

```bash
# backend ディレクトリへ移動
cd backend

# Article スキーマを作成
go run -mod=mod entgo.io/ent/cmd/ent new Article
```

これで `backend/ent/schema/article.go` が生成されます。

---

### 2. スキーマを編集

生成されたファイルにフィールドを定義します。

```go
// backend/ent/schema/article.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
)

type Article struct {
    ent.Schema
}

func (Article) Fields() []ent.Field {
    return []ent.Field{
      field.String("title"),
      field.String("body"),
    }
}

func (Article) Edges() []ent.Edge {
    return nil
}
```

---

### 3. コード生成

スキーマを元に ent のコードを生成します。

```bash
cd backend
go generate ./ent
```

`backend/ent/` 以下に `article.go`, `article_create.go`, `article_query.go` などが生成されます。

> ⚠️ `backend/ent/generate.go` に以下の記述があるので、`go generate ./ent` で自動生成できます。
>
> ```go
> //go:generate go run -mod=mod entgo.io/ent/cmd/ent generate ./schema
> ```

---

### 4. Domain の追加

`internal/domain` にシンプルなエンティティを定義します。

```go
// internal/domain/article.go
package domain

type Article struct {
    ID    int
    Title string
    Body  string
}
```

---

### 5. Repository の追加

DB とのやり取りを実装します。

```go
// internal/repository/article_repository.go
package repository

import (
    "context"

    "github.com/keu-5/muzee/backend/ent"
    "github.com/keu-5/muzee/backend/internal/domain"
)

type ArticleRepository interface {
    Create(ctx context.Context, title, body string) (*domain.Article, error)
    GetAll(ctx context.Context) ([]*domain.Article, error)
}

type articleRepository struct {
    client *ent.Client
}

func NewArticleRepository(client *ent.Client) ArticleRepository {
    return &articleRepository{client: client}
}

func (r *articleRepository) Create(ctx context.Context, title, body string) (*domain.Article, error) {
    a, err := r.client.Article.Create().
      SetTitle(title).
      SetBody(body).
      Save(ctx)
    if err != nil {
      return nil, err
    }
    return &domain.Article{ID: a.ID, Title: a.Title, Body: a.Body}, nil
}

func (r *articleRepository) GetAll(ctx context.Context) ([]*domain.Article, error) {
    articles, err := r.client.Article.Query().All(ctx)
    if err != nil {
      return nil, err
    }
    result := make([]*domain.Article, 0, len(articles))
    for _, a := range articles {
      result = append(result, &domain.Article{ID: a.ID, Title: a.Title, Body: a.Body})
    }
    return result, nil
}
```

---

### 6. Usecase の追加

ビジネスロジックを実装します。

```go
// internal/usecase/article_usecase.go
package usecase

import (
    "context"

    "github.com/keu-5/muzee/backend/internal/domain"
    "github.com/keu-5/muzee/backend/internal/repository"
)

type ArticleUsecase interface {
    CreateArticle(ctx context.Context, title, body string) (*domain.Article, error)
    GetAllArticles(ctx context.Context) ([]*domain.Article, error)
}

type articleUsecase struct {
    repo repository.ArticleRepository
}

func NewArticleUsecase(repo repository.ArticleRepository) ArticleUsecase {
    return &articleUsecase{repo: repo}
}

func (u *articleUsecase) CreateArticle(ctx context.Context, title, body string) (*domain.Article, error) {
    return u.repo.Create(ctx, title, body)
}

func (u *articleUsecase) GetAllArticles(ctx context.Context) ([]*domain.Article, error) {
    return u.repo.GetAll(ctx)
}
```

---

### 7. Handler の追加

HTTP リクエストを処理します。

```go
// internal/interface/handler/article_handler.go
package handler

import (
    "github.com/gofiber/fiber/v2"
    "github.com/keu-5/muzee/backend/internal/usecase"
)

type ArticleHandler struct {
    uc usecase.ArticleUsecase
}

func NewArticleHandler(uc usecase.ArticleUsecase) *ArticleHandler {
    return &ArticleHandler{uc: uc}
}

func (h *ArticleHandler) Create(c *fiber.Ctx) error {
    type Request struct {
      Title string `json:"title"`
      Body  string `json:"body"`
    }
    var req Request
    if err := c.BodyParser(&req); err != nil {
      return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
    }

    article, err := h.uc.CreateArticle(c.Context(), req.Title, req.Body)
    if err != nil {
      return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    return c.JSON(article)
}

func (h *ArticleHandler) GetAll(c *fiber.Ctx) error {
    articles, err := h.uc.GetAllArticles(c.Context())
    if err != nil {
      return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    return c.JSON(articles)
}
```

---

## 8. Router に登録

`internal/interface/router.go` に追加します。

```go
func RegisterRoutes(app *fiber.App, articleHandler *handler.ArticleHandler) {
    // 既存
    app.Get("/health", func(c *fiber.Ctx) error {
      return c.SendString("ok")
    })

    // 新規 Article エンドポイント
    app.Post("/articles", articleHandler.Create)
    app.Get("/articles", articleHandler.GetAll)
}
```

---

## 9. main.go に DI 設定を追加

`cmd/server/main.go` に `fx.Provide` と `fx.Invoke` を追加。

```go
fx.New(
    fx.Provide(
      // 既存
      infrastructure.NewDevelopmentLogger,
      config.Load,
      infrastructure.NewClient,
      NewFiberApp,

      // 新規 Article
      repository.NewArticleRepository,
      usecase.NewArticleUsecase,
      handler.NewArticleHandler,
    ),
    fx.Invoke(
      LogConfigLoaded,
      infrastructure.AutoMigrate,
      RegisterRoutes,
      StartServer,
    ),
).Run()
```

---

## 10. 動作確認

```bash
# dev 環境を起動
docker compose -f deploy/docker-compose.dev.yml up --build

# 動作確認
curl -X POST http://localhost:8080/articles -H "Content-Type: application/json" -d '{"title":"Hello","body":"World"}'
curl http://localhost:8080/articles
```

---

## まとめ

1. `ent init` でスキーマ作成
2. フィールド追加 → `go generate ./ent`
3. `domain` → `repository` → `usecase` → `handler` 実装
4. `router` に登録
5. `main.go` の DI に追加
6. `docker compose up` で動作確認
