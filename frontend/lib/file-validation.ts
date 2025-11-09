export const MAX_IMAGE_SIZE = 5 * 1024 * 1024; // 5MB

const allowedImageTypes = new Set([
  "image/jpeg",
  "image/png",
  "image/gif",
  "image/webp",
]);

export const base64ToFile = async (
  base64: string,
  filename: string,
): Promise<File> => {
  const res = await fetch(base64);
  const blob = await res.blob();
  return new File([blob], filename, { type: blob.type });
};

export const validateImageFile = (file: File): void => {
  if (!file) {
    throw new Error("ファイルが提供されていません");
  }

  if (file.size > MAX_IMAGE_SIZE) {
    throw new Error("ファイルサイズが大きすぎます。最大5MBまでです");
  }

  if (!allowedImageTypes.has(file.type)) {
    throw new Error(
      "サポートされていないファイル形式です。JPEG、PNG、GIF、WebPのみサポートされています",
    );
  }
};

export const getFileExtension = (contentType: string): string => {
  switch (contentType) {
    case "image/jpeg":
      return ".jpg";
    case "image/png":
      return ".png";
    case "image/gif":
      return ".gif";
    case "image/webp":
      return ".webp";
    default:
      const parts = contentType.split("/");
      if (parts.length === 2) {
        return "." + parts[1];
      }
      return "";
  }
};
