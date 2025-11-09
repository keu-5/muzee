import { AuthFormLayout } from "@/features/auth/components/auth-form-layout";
import { LogoutButton } from "@/features/auth/components/logout-button";

export default function LogoutPage() {
  return (
    <AuthFormLayout
      title="ログアウトしますか？"
      description="いつでもログインし直すことができます。アカウントを切り替える場合は、既存のアカウントを追加すると切り替えることができます。"
    >
      <LogoutButton />
    </AuthFormLayout>
  );
}
