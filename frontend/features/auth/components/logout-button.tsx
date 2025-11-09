"use client";

import { Providers } from "@/app/providers";
import { Button } from "@/components/ui/button";
import { usePostV1AuthLogout } from "@/src/api/__generated__/auth/auth";
import { useRouter } from "next/navigation";
import { toast } from "sonner";

const _LogoutButton = () => {
  const router = useRouter();

  const logoutMutation = usePostV1AuthLogout({
    mutation: {
      onSuccess: () => {
        toast("ログアウトしました");
        router.push("/login");
      },
      onError: (error) => {
        toast(error.message || "ログアウトに失敗しました");
        router.push("/login");
      },
    },
  });

  const handleLogout = () => {
    logoutMutation.mutate({
      data: {
        refresh_token: "",
      },
    });
  };

  return (
    <div className="space-y-3">
      <Button
        onClick={handleLogout}
        disabled={logoutMutation.isPending}
        className="w-full bg-foreground text-background hover:bg-foreground/90 rounded-full py-6 text-base font-medium"
      >
        {logoutMutation.isPending ? "ログアウト中..." : "ログアウト"}
      </Button>
      <Button
        onClick={router.back}
        variant="outline"
        className="w-full rounded-full py-6 text-base font-medium"
      >
        キャンセル
      </Button>
    </div>
  );
};

export const LogoutButton = () => {
  return (
    <Providers>
      <_LogoutButton />
    </Providers>
  );
};
