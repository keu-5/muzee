"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

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
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSlot,
} from "@/components/ui/input-otp";
import { Label } from "@/components/ui/label";
import {
  usePostV1AuthSignupResendCode,
  usePostV1AuthSignupVerifyCode,
} from "@/src/api/__generated__/auth/auth";
import { toast } from "sonner";

const verifySchema = z.object({
  code: z
    .string()
    .length(6, "6桁の認証コードを入力してください")
    .regex(/^\d{6}$/, "数字のみで入力してください"),
});

type VerifyFormValues = z.infer<typeof verifySchema>;

const _VerifyCodeForm = () => {
  const router = useRouter();
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [email, setEmail] = useState("");
  const { mutate: verifyCode } = usePostV1AuthSignupVerifyCode();
  const { mutate: resendCode } = usePostV1AuthSignupResendCode();

  const form = useForm<VerifyFormValues>({
    resolver: zodResolver(verifySchema),
    defaultValues: { code: "" },
  });

  const code = form.watch("code");

  useEffect(() => {
    const storedEmail = sessionStorage.getItem("signupEmail") || "";
    setEmail(storedEmail);
  }, []);

  const handleVerifyCode = async (data: VerifyFormValues) => {
    setError("");
    setIsLoading(true);

    verifyCode(
      {
        data: {
          client_id: process.env.NEXT_PUBLIC_CLIENT_ID || "",
          email,
          code: data.code,
        },
      },
      {
        onSuccess: () => {
          sessionStorage.removeItem("signupEmail");
          router.push("/"); //TODO: redirect to home
        },
        onError: (err) => {
          setError(err.message || "アカウントの作成に失敗しました");
        },
        onSettled: () => {
          setIsLoading(false);
        },
      },
    );
  };

  const handleResendCode = async () => {
    setError("");
    setIsLoading(true);

    resendCode(
      { data: { email } },
      {
        onSuccess: (res) => {
          toast(res.message || "認証コードを再送信しました！");
        },
        onError: (err) => {
          setError(err.message || "認証コードの再送信に失敗しました");
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
          メール認証
        </CardTitle>
        <CardDescription className="text-center">
          メールに送信された6桁の認証コードを入力してください
        </CardDescription>
      </CardHeader>

      <form onSubmit={form.handleSubmit(handleVerifyCode)}>
        <CardContent className="space-y-4 my-6">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <div className="space-y-2">
            <Label htmlFor="code" className="text-center block">
              認証コード
            </Label>
            <div className="flex justify-center">
              <InputOTP
                maxLength={6}
                {...form.register("code")}
                value={code}
                onChange={(value) => form.setValue("code", value)}
                disabled={isLoading}
              >
                <InputOTPGroup>
                  {[0, 1, 2, 3, 4, 5].map((i) => (
                    <InputOTPSlot key={i} index={i} />
                  ))}
                </InputOTPGroup>
              </InputOTP>
            </div>

            {form.formState.errors.code && (
              <p className="text-sm text-red-500 text-center">
                {form.formState.errors.code.message}
              </p>
            )}

            <p className="text-xs text-muted-foreground text-center">
              {email ? `${email} に送信されました` : "メールアドレス未設定"}
            </p>
          </div>
        </CardContent>

        <CardFooter className="flex flex-col space-y-3">
          <Button
            type="submit"
            className="w-full"
            disabled={isLoading || code.length !== 6}
          >
            {isLoading ? "確認中..." : "アカウントを作成"}
          </Button>

          <Button
            type="button"
            variant="outline"
            className="w-full bg-transparent"
            onClick={handleResendCode}
            disabled={isLoading}
          >
            認証コードを再送信
          </Button>

          <Button
            type="button"
            variant="ghost"
            className="w-full"
            onClick={() => router.push("/signup")}
            disabled={isLoading}
          >
            戻る
          </Button>
        </CardFooter>
      </form>
    </Card>
  );
};

export const VerifyCodeForm = () => {
  return (
    <Providers>
      <_VerifyCodeForm />
    </Providers>
  );
};
