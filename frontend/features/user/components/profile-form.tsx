"use client";

import { Providers } from "@/app/providers";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ImageCropModal } from "@/features/user/components/image-crop-modal";
import { LINK } from "@/lib/links";
import { useGetV1UserProfilesCheckUsername } from "@/src/api/__generated__/user-profiles/user-profiles";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  AtSign,
  CheckCircle2,
  Cloud,
  Pencil,
  Trash2,
  User,
  XCircle,
} from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { useDebounce } from "use-debounce";
import { z } from "zod";

const profileSchema = z.object({
  name: z.string().trim().min(1, "名前を入力してください"),
  username: z
    .string()
    .trim()
    .min(3, "ユーザー名は3文字以上である必要があります"),
});

type ProfileFormValues = z.infer<typeof profileSchema>;

const _ProfileForm = () => {
  const router = useRouter();
  const [iconImage, setIconImage] = useState<string>("");
  const [originalImage, setOriginalImage] = useState<string>("");
  const [cropParams, setCropParams] = useState<{
    x: number;
    y: number;
    size: number;
  } | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [tempImageSrc, setTempImageSrc] = useState<string>("");
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
    setError: setFormError,
    clearErrors,
  } = useForm<ProfileFormValues>({
    resolver: zodResolver(profileSchema),
    mode: "onChange",
  });

  const username = watch("username");
  const [debouncedUsername] = useDebounce(username, 500);

  const { data: usernameCheckData, isFetching: isCheckingUsername } =
    useGetV1UserProfilesCheckUsername(
      { username: debouncedUsername || "" },
      {
        query: {
          enabled:
            !!debouncedUsername &&
            debouncedUsername.length >= 3 &&
            !errors.username,
        },
      },
    );

  useEffect(() => {
    if (!debouncedUsername || debouncedUsername.length < 3) {
      clearErrors("username");
      return;
    }

    if (usernameCheckData && !usernameCheckData.available) {
      setFormError("username", {
        type: "manual",
        message: "このユーザー名は既に使用されています",
      });
    } else if (usernameCheckData?.available) {
      clearErrors("username");
    }
  }, [usernameCheckData, debouncedUsername, setFormError, clearErrors]);

  const handleDragOver = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.currentTarget.classList.add("bg-accent");
  };

  const handleDragLeave = (e: React.DragEvent<HTMLDivElement>) => {
    e.currentTarget.classList.remove("bg-accent");
  };

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.currentTarget.classList.remove("bg-accent");
    const file = e.dataTransfer.files?.[0];
    if (file) processImageFile(file);
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) processImageFile(file);
    e.target.value = "";
  };

  const processImageFile = (file: File) => {
    const reader = new FileReader();
    reader.onload = (event) => {
      const src = event.target?.result as string;
      setOriginalImage(src);
      setTempImageSrc(src);
      setCropParams(null);
      setIsModalOpen(true);
    };
    reader.readAsDataURL(file);
  };

  const handleImageSave = (
    croppedImage: string,
    params: { x: number; y: number; size: number },
  ) => {
    setIconImage(croppedImage);
    setCropParams(params);
  };

  const onSubmit = async (data: ProfileFormValues) => {
    setError("");
    setIsLoading(true);

    try {
      //TODO: モックAPIコール
      await new Promise((resolve) => setTimeout(resolve, 1000));

      const profileData = {
        icon_path: iconImage,
        name: data.name,
        username: data.username,
        createdAt: new Date().toISOString(),
      };

      localStorage.setItem("userProfile", JSON.stringify(profileData));

      router.push(LINK.home);
    } catch (err) {
      setError("プロフィール作成に失敗しました。");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <>
      <Card className="w-full max-w-md shadow-xl">
        <CardHeader className="space-y-1">
          <CardTitle className="text-2xl font-bold text-center">
            プロフィール設定
          </CardTitle>
          <CardDescription className="text-center">
            あなたのプロフィール情報を入力してください
          </CardDescription>
        </CardHeader>

        <form onSubmit={handleSubmit(onSubmit)} noValidate>
          <CardContent className="space-y-4">
            {error && (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            {/* 画像アップロード */}
            <div className="space-y-2">
              <Label htmlFor="icon">プロフィール画像</Label>
              {iconImage ? (
                <div className="relative group mx-auto w-32 h-32 rounded-full overflow-hidden border-2 border-[#BDB76B] shadow-sm transition-all duration-200 hover:shadow-md">
                  <img
                    src={iconImage}
                    alt="Profile"
                    className="w-full h-full object-cover"
                  />
                  <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 flex items-center justify-center gap-3 transition-opacity">
                    {/* 編集ボタン */}
                    <Button
                      size="icon"
                      variant="secondary"
                      className="bg-white/80 hover:bg-white text-black rounded-full"
                      onClick={() => {
                        setTempImageSrc(originalImage);
                        setIsModalOpen(true);
                      }}
                      type="button"
                    >
                      <Pencil className="h-4 w-4" />
                    </Button>

                    {/* 削除ボタン */}
                    <Button
                      size="icon"
                      variant="secondary"
                      className="bg-white/80 hover:bg-red-100 text-red-600 rounded-full"
                      onClick={() => {
                        setIconImage("");
                        setOriginalImage("");
                        setCropParams(null);
                      }}
                      type="button"
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              ) : (
                <div
                  onDragOver={handleDragOver}
                  onDragLeave={handleDragLeave}
                  onDrop={handleDrop}
                  className="border-2 border-dashed border-[#BDB76B] rounded-lg p-6 text-center transition-colors duration-200 cursor-pointer hover:bg-accent/50"
                >
                  <input
                    id="icon-input"
                    type="file"
                    accept="image/*"
                    onChange={handleFileChange}
                    disabled={isLoading}
                    className="hidden"
                  />
                  <label
                    htmlFor="icon-input"
                    className="flex flex-col items-center gap-2 cursor-pointer"
                  >
                    <Cloud className="h-8 w-8 text-[#BDB76B]" />
                    <p className="text-sm font-medium">
                      画像をドラッグ＆ドロップ
                    </p>
                    <p className="text-xs text-muted-foreground">
                      またはクリックして選択
                    </p>
                  </label>
                </div>
              )}
            </div>

            {/* 名前 */}
            <div className="space-y-2">
              <Label htmlFor="name">名前</Label>
              <div className="relative">
                <User className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  id="name"
                  type="text"
                  placeholder="山田太郎"
                  disabled={isLoading}
                  {...register("name")}
                  className="pl-10"
                />
              </div>
              {errors.name && (
                <p className="text-sm text-red-500">{errors.name.message}</p>
              )}
            </div>

            {/* ユーザー名 */}
            <div className="space-y-2">
              <Label htmlFor="username">ユーザー名</Label>
              <div className="relative">
                <AtSign className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  id="username"
                  type="text"
                  placeholder="yamada_taro"
                  disabled={isLoading}
                  {...register("username")}
                  className="pl-10 pr-10"
                />
                {username && username.length >= 3 && (
                  <div className="absolute right-3 top-1/2 -translate-y-1/2">
                    {isCheckingUsername ? (
                      <div className="h-4 w-4 animate-spin rounded-full border-2 border-muted-foreground border-t-transparent" />
                    ) : usernameCheckData?.available ? (
                      <CheckCircle2 className="h-4 w-4 text-green-500" />
                    ) : (
                      <XCircle className="h-4 w-4 text-red-500" />
                    )}
                  </div>
                )}
              </div>
              {errors.username && (
                <p className="text-sm text-red-500">
                  {errors.username.message}
                </p>
              )}
              {!errors.username &&
                usernameCheckData?.available &&
                username &&
                username.length >= 3 && (
                  <p className="text-sm text-green-600">
                    このユーザー名は利用可能です
                  </p>
                )}
            </div>
          </CardContent>

          <CardFooter className="flex flex-col my-6">
            <Button type="submit" className="w-full" disabled={isLoading}>
              {isLoading ? "作成中..." : "プロフィールを作成"}
            </Button>
          </CardFooter>
        </form>
      </Card>

      <ImageCropModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSave={handleImageSave}
        imageSrc={tempImageSrc}
        initialCropParams={cropParams}
      />
    </>
  );
};

export const ProfileForm = () => {
  return (
    <Providers>
      <_ProfileForm />
    </Providers>
  );
};
