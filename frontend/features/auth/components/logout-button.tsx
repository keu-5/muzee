"use client";

import { Providers } from "@/app/providers";
import { Button } from "@/components/ui/button";
import { usePostV1AuthLogout } from "@/src/api/__generated__/auth/auth";
import { LogOut } from "lucide-react";
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
    <Button
      onClick={handleLogout}
      disabled={logoutMutation.isPending}
      variant="ghost"
      size="sm"
      className="gap-2"
    >
      <LogOut className="h-4 w-4" />
      {logoutMutation.isPending ? "ログアウト中..." : "ログアウト"}
    </Button>
  );
};

export const LogoutButton = () => {
  return (
    <Providers>
      <_LogoutButton />
    </Providers>
  );
};
