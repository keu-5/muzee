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
import { LINK } from "@/lib/links";
import { usePostV1AuthSignupSendCode } from "@/src/api/__generated__/auth/auth";
import { zodResolver } from "@hookform/resolvers/zod";
import { Lock, Mail } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

const COMMON_WEAK = [
  "password",
  "password1",
  "qwerty",
  "letmein",
  "12345678",
  "123456789",
  "iloveyou",
];

const signupSchema = z
  .object({
    email: z
      .email("正しいメールアドレスを入力してください")
      .trim()
      .min(1, "メールアドレスを入力してください")
      .max(254, "メールアドレスが長すぎます"),
    password: z
      .string()
      .min(8, "パスワードは8文字以上で入力してください")
      .max(64, "パスワードは64文字以内で入力してください"),
    confirmPassword: z.string().min(1, "確認用パスワードを入力してください"),
  })
  .superRefine((data, ctx) => {
    const pwd = data.password;

    // 空白（スペースや改行）を禁止
    if (/\s/.test(pwd)) {
      ctx.addIssue({
        code: "custom",
        path: ["password"],
        message: "空白文字は使用できません",
      });
    }

    // 英字 / 数字 を1文字以上
    if (!/[a-z]/.test(pwd)) {
      ctx.addIssue({
        code: "custom",
        path: ["password"],
        message: "英字（a–z）を1文字以上含めてください",
      });
    }
    if (!/[0-9]/.test(pwd)) {
      ctx.addIssue({
        code: "custom",
        path: ["password"],
        message: "数字（0–9）を1文字以上含めてください",
      });
    }

    // よくある弱いパスワードの簡易ブロック
    const lowered = pwd.toLowerCase();
    if (COMMON_WEAK.includes(lowered)) {
      ctx.addIssue({
        code: "custom",
        path: ["password"],
        message: "推測されやすいパスワードは使用できません",
      });
    }

    // email とパスワードの包含を防止
    const emailLocalPart = data.email.split("@")[0]?.toLowerCase();
    if (
      emailLocalPart &&
      lowered.includes(emailLocalPart) &&
      emailLocalPart.length >= 3
    ) {
      ctx.addIssue({
        code: "custom",
        path: ["password"],
        message: "メールアドレスに由来する文字列を含めないでください",
      });
    }

    // 確認用パスワード一致チェック
    if (data.password !== data.confirmPassword) {
      ctx.addIssue({
        code: "custom",
        path: ["confirmPassword"],
        message: "パスワードが一致しません",
      });
    }
  });

type SignupFormValues = z.infer<typeof signupSchema>;

const _SignupForm = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const { mutate: sendCode } = usePostV1AuthSignupSendCode();
  const router = useRouter();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<SignupFormValues>({
    resolver: zodResolver(signupSchema),
    mode: "onBlur",
  });

  const onSubmit = (data: SignupFormValues) => {
    setIsLoading(true);
    setError("");

    sendCode(
      { data: { email: data.email, password: data.password } },
      {
        onSuccess: (res) => {
          toast(res.message || "認証コードを送信しました！");
          router.push(LINK.signup.verify);
        },
        onError: (err) => {
          setError(
            err.message || "登録に失敗しました。もう一度お試しください。",
          );
        },
        onSettled: () => {
          setIsLoading(false);
        },
      },
    );
  };

  return (
    <Card className="w-full max-w-md shadow-xl">
      <CardHeader className="space-y-1">
        <CardTitle className="text-2xl font-bold text-center">
          新規登録
        </CardTitle>
        <CardDescription className="text-center">
          メールアドレスとパスワードを入力してください
        </CardDescription>
      </CardHeader>

      <form onSubmit={handleSubmit(onSubmit)} noValidate>
        <CardContent className="space-y-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <div className="space-y-2">
            <Label htmlFor="email">メールアドレス</Label>
            <div className="relative">
              <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                id="email"
                type="email"
                placeholder="example@email.com"
                disabled={isLoading}
                {...register("email")}
                className="pl-10"
                autoComplete="email"
                inputMode="email"
              />
            </div>
            {errors.email && (
              <p className="text-sm text-red-500">{errors.email.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="password">パスワード</Label>
            <div className="relative">
              <Lock className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                id="password"
                type="password"
                placeholder="••••••••"
                disabled={isLoading}
                {...register("password")}
                className="pl-10"
                autoComplete="new-password"
              />
            </div>
            {errors.password && (
              <ul className="mt-1 space-y-0.5 text-sm text-red-500">
                <li>{errors.password.message}</li>
              </ul>
            )}
            <p className="text-xs text-muted-foreground">
              8–64文字・半角英数字を各1文字以上・空白不可
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="confirmPassword">パスワード（確認）</Label>
            <div className="relative">
              <Lock className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                id="confirmPassword"
                type="password"
                placeholder="••••••••"
                disabled={isLoading}
                {...register("confirmPassword")}
                className="pl-10"
                autoComplete="new-password"
              />
            </div>
            {errors.confirmPassword && (
              <p className="text-sm text-red-500">
                {errors.confirmPassword.message}
              </p>
            )}
          </div>
        </CardContent>

        <CardFooter className="flex flex-col space-y-4 my-6">
          <Button type="submit" className="w-full" disabled={isLoading}>
            {isLoading ? "送信中..." : "認証コードを送信"}
          </Button>
          <p className="text-sm text-center text-muted-foreground">
            すでにアカウントをお持ちの方は{" "}
            <Link
              href={LINK.login}
              className="text-primary hover:underline font-medium"
            >
              ログイン
            </Link>
          </p>
        </CardFooter>
      </form>
    </Card>
  );
};

export const SignupForm = () => {
  return (
    <Providers>
      <_SignupForm />
    </Providers>
  );
};
