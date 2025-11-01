import { AuthFormLayout } from "@/features/auth/components/auth-form-layout";
import { VerifyCodeForm } from "@/features/auth/components/verify-code-form";

export default function VerifyPage() {
  return (
    <AuthFormLayout
      title="Verify Your Email"
      description="認証コードを入力してアカウントを作成"
    >
      <VerifyCodeForm />
    </AuthFormLayout>
  );
}
