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
import { usePostV1AuthLogin } from "@/src/api/__generated__/auth/auth";
import { zodResolver } from "@hookform/resolvers/zod";
import { Lock, Mail } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useForm } from "react-hook-form";
import z from "zod";

const loginSchema = z.object({
  email: z
    .email("正しいメールアドレスを入力してください")
    .trim()
    .min(1, "メールアドレスを入力してください"),
  password: z.string(),
});

type LoginFormValues = z.infer<typeof loginSchema>;

const _LoginForm = () => {
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const { mutate: login } = usePostV1AuthLogin();
  const router = useRouter();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    mode: "onBlur",
  });

  const onSubmit = (data: LoginFormValues) => {
    setError("");
    setIsLoading(true);

    login(
      {
        data: {
          email: data.email,
          password: data.password,
          client_id: process.env.NEXT_PUBLIC_CLIENT_ID || "",
        },
      },
      {
        onSuccess: () => {
          router.push(LINK.home);
        },
        onError: (err) => {
          setError(
            err.message || "ログインに失敗しました。もう一度お試しください。",
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
          ログイン
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
                autoComplete="email"
                inputMode="email"
                className="pl-10"
              />
            </div>
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
                autoComplete="current-password"
              />
            </div>
          </div>
          {errors.password && (
            <ul className="mt-1 space-y-0.5 text-sm text-red-500">
              <li>{errors.password.message}</li>
            </ul>
          )}
        </CardContent>

        <CardFooter className="flex flex-col space-y-4">
          <Button type="submit" className="w-full" disabled={isLoading}>
            {isLoading ? "ログイン中..." : "ログイン"}
          </Button>
          <p className="text-sm text-center text-muted-foreground">
            アカウントをお持ちでない方は{" "}
            <Link
              href={LINK.signup.base}
              className="text-primary hover:underline font-medium"
            >
              新規登録
            </Link>
          </p>
        </CardFooter>
      </form>
    </Card>
  );
};

export const LoginForm = () => {
  return (
    <Providers>
      <_LoginForm />
    </Providers>
  );
};
