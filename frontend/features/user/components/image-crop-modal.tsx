"use client";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { X } from "lucide-react";
import { useEffect, useRef, useState } from "react";

interface ImageCropModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (
    croppedImage: string,
    params: { x: number; y: number; size: number },
  ) => void;
  imageSrc: string;
  initialCropParams?: { x: number; y: number; size: number } | null;
}

export function ImageCropModal({
  isOpen,
  onClose,
  onSave,
  imageSrc,
  initialCropParams,
}: ImageCropModalProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const imageRef = useRef<HTMLImageElement>(null);
  const [isDraggingMove, setIsDraggingMove] = useState(false);
  const [isDraggingResize, setIsDraggingResize] = useState(false);
  const [resizeCorner, setResizeCorner] = useState<string>("");
  const [lastX, setLastX] = useState(0);
  const [lastY, setLastY] = useState(0);
  const [canvasWidth, setCanvasWidth] = useState(400);
  const [canvasHeight, setCanvasHeight] = useState(400);
  const [cropX, setCropX] = useState(100);
  const [cropY, setCropY] = useState(100);
  const [cropSize, setCropSize] = useState(200);

  const MIN_CROP_SIZE = 50;
  const CORNER_SIZE = 30; // コーナーのヒット判定サイズ

  // 画像読み込み時にキャンバスサイズを設定
  useEffect(() => {
    const img = imageRef.current;
    if (!img || !imageSrc) return;

    const handleLoad = () => {
      // 画像のアスペクト比を維持しながら、最大幅を設定
      const maxWidth = 400;
      const maxHeight = 400;
      let width = img.naturalWidth;
      let height = img.naturalHeight;

      if (width > maxWidth || height > maxHeight) {
        const ratio = Math.min(maxWidth / width, maxHeight / height);
        width = width * ratio;
        height = height * ratio;
      }

      setCanvasWidth(width);
      setCanvasHeight(height);

      // 初期クロップパラメータがあればそれを使用、なければデフォルト
      if (initialCropParams) {
        setCropX(initialCropParams.x);
        setCropY(initialCropParams.y);
        setCropSize(initialCropParams.size);
      } else {
        const initialSize = Math.min(width, height) * 0.5;
        setCropSize(initialSize);
        setCropX((width - initialSize) / 2);
        setCropY((height - initialSize) / 2);
      }
    };

    if (img.complete) {
      handleLoad();
    } else {
      img.addEventListener("load", handleLoad);
      return () => img.removeEventListener("load", handleLoad);
    }
  }, [imageSrc, isOpen, initialCropParams]);

  // 画像を描画
  useEffect(() => {
    const canvas = canvasRef.current;
    const img = imageRef.current;
    if (!canvas || !img || !img.complete) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    // Clear canvas
    ctx.clearRect(0, 0, canvasWidth, canvasHeight);

    // Draw the full image
    ctx.drawImage(img, 0, 0, canvasWidth, canvasHeight);

    // Draw overlay mask (darken area outside crop box)
    ctx.fillStyle = "rgba(0,0,0,0.5)";
    ctx.fillRect(0, 0, canvasWidth, canvasHeight);

    // Clear the crop area to show the image
    ctx.clearRect(cropX, cropY, cropSize, cropSize);

    // Redraw image in the crop area
    ctx.drawImage(img, 0, 0, canvasWidth, canvasHeight);

    // Draw grid lines (3x3 grid) - 三分割法のガイドライン
    ctx.strokeStyle = "rgba(255,255,255,0.5)";
    ctx.lineWidth = 1;
    ctx.setLineDash([5, 5]);

    // Vertical lines
    for (let i = 1; i < 3; i++) {
      const x = cropX + (cropSize / 3) * i;
      ctx.beginPath();
      ctx.moveTo(x, cropY);
      ctx.lineTo(x, cropY + cropSize);
      ctx.stroke();
    }

    // Horizontal lines
    for (let i = 1; i < 3; i++) {
      const y = cropY + (cropSize / 3) * i;
      ctx.beginPath();
      ctx.moveTo(cropX, y);
      ctx.lineTo(cropX + cropSize, y);
      ctx.stroke();
    }

    ctx.setLineDash([]);

    // Draw crop box border - クロップ枠の装飾
    // 外側の線 (太め)
    ctx.strokeStyle = "rgba(189, 183, 107, 0.8)";
    ctx.lineWidth = 3;
    ctx.strokeRect(cropX, cropY, cropSize, cropSize);

    // 内側の線 (細め)
    ctx.strokeStyle = "rgba(255, 255, 255, 0.9)";
    ctx.lineWidth = 1;
    ctx.strokeRect(cropX + 1.5, cropY + 1.5, cropSize - 3, cropSize - 3);

    // Draw corner handles - コーナーハンドル
    const corners = [
      { x: cropX, y: cropY, position: "tl" }, // 左上
      { x: cropX + cropSize, y: cropY, position: "tr" }, // 右上
      { x: cropX, y: cropY + cropSize, position: "bl" }, // 左下
      { x: cropX + cropSize, y: cropY + cropSize, position: "br" }, // 右下
    ];

    const cornerLength = 20;
    const cornerWidth = 3;

    ctx.strokeStyle = "#BDB76B";
    ctx.lineWidth = cornerWidth;
    ctx.lineCap = "round";

    corners.forEach((corner) => {
      // L字型のコーナーハンドル
      if (corner.position === "tl") {
        // 左上: 横線と縦線
        ctx.beginPath();
        ctx.moveTo(corner.x, corner.y + cornerLength);
        ctx.lineTo(corner.x, corner.y);
        ctx.lineTo(corner.x + cornerLength, corner.y);
        ctx.stroke();
      } else if (corner.position === "tr") {
        // 右上
        ctx.beginPath();
        ctx.moveTo(corner.x - cornerLength, corner.y);
        ctx.lineTo(corner.x, corner.y);
        ctx.lineTo(corner.x, corner.y + cornerLength);
        ctx.stroke();
      } else if (corner.position === "bl") {
        // 左下
        ctx.beginPath();
        ctx.moveTo(corner.x, corner.y - cornerLength);
        ctx.lineTo(corner.x, corner.y);
        ctx.lineTo(corner.x + cornerLength, corner.y);
        ctx.stroke();
      } else if (corner.position === "br") {
        // 右下
        ctx.beginPath();
        ctx.moveTo(corner.x - cornerLength, corner.y);
        ctx.lineTo(corner.x, corner.y);
        ctx.lineTo(corner.x, corner.y - cornerLength);
        ctx.stroke();
      }
    });
  }, [canvasWidth, canvasHeight, cropX, cropY, cropSize, isOpen]);

  // クリック位置がコーナーかどうかを判定
  const getCornerAtPosition = (x: number, y: number): string => {
    const canvas = canvasRef.current;
    if (!canvas) return "";

    const rect = canvas.getBoundingClientRect();
    const scaleX = canvasWidth / rect.width;
    const scaleY = canvasHeight / rect.height;
    const canvasX = (x - rect.left) * scaleX;
    const canvasY = (y - rect.top) * scaleY;

    const corners = [
      { name: "top-left", x: cropX, y: cropY },
      { name: "top-right", x: cropX + cropSize, y: cropY },
      { name: "bottom-left", x: cropX, y: cropY + cropSize },
      { name: "bottom-right", x: cropX + cropSize, y: cropY + cropSize },
    ];

    for (const corner of corners) {
      if (
        Math.abs(canvasX - corner.x) < CORNER_SIZE &&
        Math.abs(canvasY - corner.y) < CORNER_SIZE
      ) {
        return corner.name;
      }
    }

    return "";
  };

  const handlePointerDown = (e: React.PointerEvent) => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const rect = canvas.getBoundingClientRect();
    const scaleX = canvasWidth / rect.width;
    const scaleY = canvasHeight / rect.height;
    const canvasX = (e.clientX - rect.left) * scaleX;
    const canvasY = (e.clientY - rect.top) * scaleY;

    // コーナーをクリックしたか確認
    const corner = getCornerAtPosition(e.clientX, e.clientY);
    if (corner) {
      setIsDraggingResize(true);
      setResizeCorner(corner);
      setLastX(e.clientX);
      setLastY(e.clientY);
      e.currentTarget.setPointerCapture(e.pointerId);
      return;
    }

    // 枠内をクリックしたか確認
    if (
      canvasX >= cropX &&
      canvasX <= cropX + cropSize &&
      canvasY >= cropY &&
      canvasY <= cropY + cropSize
    ) {
      setIsDraggingMove(true);
      setLastX(e.clientX);
      setLastY(e.clientY);
      e.currentTarget.setPointerCapture(e.pointerId);
    }
  };

  const handlePointerMove = (e: React.PointerEvent) => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const rect = canvas.getBoundingClientRect();
    const scaleX = canvasWidth / rect.width;
    const scaleY = canvasHeight / rect.height;
    const dx = (e.clientX - lastX) * scaleX;
    const dy = (e.clientY - lastY) * scaleY;

    if (isDraggingResize) {
      // リサイズ処理
      let newX = cropX;
      let newY = cropY;
      let newSize = cropSize;

      // 正方形を維持するため、dxとdyの平均を取る
      const delta = (dx + dy) / 2;

      if (resizeCorner === "top-left") {
        // 左上: サイズを小さくする方向に移動
        newSize = cropSize - delta;
        newX = cropX + delta;
        newY = cropY + delta;
      } else if (resizeCorner === "top-right") {
        // 右上: 右に動かすとサイズ増加、上に動かすとサイズ減少
        const deltaResize = (dx - dy) / 2;
        newSize = cropSize + deltaResize;
        newY = cropY - deltaResize;
      } else if (resizeCorner === "bottom-left") {
        // 左下: 左に動かすとサイズ減少、下に動かすとサイズ増加
        const deltaResize = (-dx + dy) / 2;
        newSize = cropSize + deltaResize;
        newX = cropX - deltaResize;
      } else if (resizeCorner === "bottom-right") {
        // 右下: 両方向に拡大
        newSize = cropSize + delta;
      }

      // 最小サイズと画像範囲内に制限
      newSize = Math.max(MIN_CROP_SIZE, newSize);

      // 画像範囲内に収まるように調整
      if (newX < 0) {
        newSize = newSize + newX;
        newX = 0;
      }
      if (newY < 0) {
        newSize = newSize + newY;
        newY = 0;
      }
      if (newX + newSize > canvasWidth) {
        newSize = canvasWidth - newX;
      }
      if (newY + newSize > canvasHeight) {
        newSize = canvasHeight - newY;
      }

      // サイズが最小値以上の場合のみ更新
      if (newSize >= MIN_CROP_SIZE && newX >= 0 && newY >= 0) {
        setCropX(newX);
        setCropY(newY);
        setCropSize(newSize);
        setLastX(e.clientX);
        setLastY(e.clientY);
      }
    } else if (isDraggingMove) {
      // 移動処理
      const newX = Math.max(0, Math.min(canvasWidth - cropSize, cropX + dx));
      const newY = Math.max(0, Math.min(canvasHeight - cropSize, cropY + dy));

      setCropX(newX);
      setCropY(newY);
      setLastX(e.clientX);
      setLastY(e.clientY);
    } else {
      // カーソル変更
      const corner = getCornerAtPosition(e.clientX, e.clientY);
      if (canvas) {
        if (corner === "top-left" || corner === "bottom-right") {
          canvas.style.cursor = "nwse-resize";
        } else if (corner === "top-right" || corner === "bottom-left") {
          canvas.style.cursor = "nesw-resize";
        } else {
          canvas.style.cursor = "move";
        }
      }
    }
  };

  const handlePointerUp = (e: React.PointerEvent) => {
    setIsDraggingMove(false);
    setIsDraggingResize(false);
    setResizeCorner("");
    e.currentTarget.releasePointerCapture(e.pointerId);
  };

  const handleSave = () => {
    const canvas = canvasRef.current;
    const img = imageRef.current;
    if (!canvas || !img) return;

    const cropCanvas = document.createElement("canvas");
    cropCanvas.width = cropSize;
    cropCanvas.height = cropSize;
    const cropCtx = cropCanvas.getContext("2d");
    if (!cropCtx) return;

    // 元画像から切り抜き領域を計算
    const scaleX = img.naturalWidth / canvasWidth;
    const scaleY = img.naturalHeight / canvasHeight;

    cropCtx.drawImage(
      img,
      cropX * scaleX,
      cropY * scaleY,
      cropSize * scaleX,
      cropSize * scaleY,
      0,
      0,
      cropSize,
      cropSize,
    );

    onSave(cropCanvas.toDataURL("image/png"), {
      x: cropX,
      y: cropY,
      size: cropSize,
    });
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
      onClick={(e) => {
        if (e.target === e.currentTarget) {
          onClose();
        }
      }}
    >
      <Card
        className="w-full max-w-lg p-6 bg-background"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-bold">画像をクロップ</h2>
          <button
            onClick={onClose}
            className="text-muted-foreground hover:text-foreground"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        <p className="text-sm text-muted-foreground mb-2">
          枠をドラッグして移動、四隅をドラッグしてサイズ変更できます。
        </p>

        <div className="flex justify-center">
          <canvas
            ref={canvasRef}
            width={canvasWidth}
            height={canvasHeight}
            className="border-2 border-[#BDB76B] rounded-lg touch-none max-w-full h-auto"
            onPointerDown={handlePointerDown}
            onPointerMove={handlePointerMove}
            onPointerUp={handlePointerUp}
            onPointerCancel={handlePointerUp}
          />
        </div>

        <img
          ref={imageRef}
          src={imageSrc || "/placeholder.svg"}
          alt="source"
          className="hidden"
          crossOrigin="anonymous"
        />

        <div className="flex gap-2 mt-4">
          <Button
            variant="outline"
            onClick={onClose}
            className="flex-1 bg-transparent"
          >
            キャンセル
          </Button>
          <Button
            onClick={handleSave}
            className="flex-1 bg-[#BDB76B] text-black hover:bg-[#A0A060]"
          >
            決定
          </Button>
        </div>
      </Card>
    </div>
  );
}
