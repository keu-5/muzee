#!/bin/sh
set -e

echo "⏳ MinIO バケット初期化中..."

# MinIO 起動を少し待つ
sleep 5

# 接続設定
mc alias set local "$S3_ENDPOINT" "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD"

# バケット作成（すでに存在する場合はスキップ）
mc mb -p local/public-uploads || true
mc mb -p local/private-uploads || true

# 公開設定
mc anonymous set download local/public-uploads || true
mc anonymous set none local/private-uploads || true

echo "✅ MinIO バケット初期化完了！"

exit 0