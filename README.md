# muzee

## 必要環境

- Docker / Docker Compose（推奨）

## 構成

- **Docker / Docker Compose**
  各サービス（バックエンド / フロントエンド / DB / Nginx）をコンテナとして起動・管理します。開発・本番環境の切り替えも容易に行えます。

- **Go**
  バックエンド API サーバー。軽量フレームワーク **Fiber** を利用し、**Air** によるホットリロードをサポートしています。
  → 詳細は [backend/README.md](backend/README.md)

- **Next.js**
  フロントエンドアプリケーション。React ベースで SSR/SSG に対応し、スタイリングには **Tailwind CSS** を利用しています。
  → 詳細は [frontend/README.md](frontend/README.md)

- **PostgreSQL**
  メインのリレーショナルデータベース。ユーザー情報やアプリケーションデータを永続化します。

- **Nginx**
  リバースプロキシとして利用。フロントエンド（Next.js）とバックエンド（Go API）へのリクエストをルーティングします。

## 開発環境構築

### プロジェクトの取り込み

```zsh
git clone https://github.com/keu-5/muzee.git
```

### docker 環境構築

```zsh
cd deploy
```

```zsh
make dev-build
```

### docker 実行

```zsh
make dev-up
```

`localhost`, `localhost/api/`にアクセス可能になります。

```zsh
make dev-down
```

## api クライアント作成

バックエンドで各 handler に swaggo 用のコメントアウトを書き、以下を実行するとフロントエンドで api クライアントが生成される

```shell
make gen-all
```

コメントアウトの書き方は実装を確認

docker compose 起動後、`http://localhost/api/docs/index.html`で確認可能
