# Frontend

## 概要

このディレクトリは **Next.js** を利用したフロントエンドアプリケーションです。
スタイルには **Tailwind CSS** を使用し、パッケージマネージャには **pnpm** を採用しています。

---

## 構成

- **[Next.js](https://nextjs.org/)**
  React ベースのフレームワーク。SSR/SSG に対応しており、高速なレンダリングと SEO 対策が可能です。

- **[Tailwind CSS](https://tailwindcss.com/)**
  ユーティリティファーストの CSS フレームワーク。効率的にレスポンシブかつモダンなデザインを実装できます。

- **[pnpm](https://pnpm.io/)**
  高速でディスク効率の良いパッケージマネージャ。依存関係の管理とインストールを最適化します。

- **[orval](https://orval.dev/)**
  OpenAPI 仕様から自動的に型安全な API クライアントを生成するツール。
  React Query と組み合わせることで、API 通信の型保証と効率的なキャッシュ管理を実現します。

- **[TanStack Query](https://tanstack.com/query/v4)**
  データフェッチング・キャッシュ・状態管理を簡潔に扱えるライブラリ。
  orval で生成したクライアントと組み合わせ、効率的かつ宣言的な API データ操作が可能です。

- **[Axios](https://axios-http.com/)**
  Promise ベースの HTTP クライアント。
  orval の出力にカスタムラッパーを組み込むことで、SSR/CSR 双方でのリクエスト処理を統一できます。
