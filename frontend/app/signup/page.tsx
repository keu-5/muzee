import { AuthFormLayout } from "@/features/auth/components/auth-form-layout";
import { SignupForm } from "@/features/auth/components/signup-form";

export default function SignupPage() {
  return (
    <AuthFormLayout
      title="Get Started"
      description="新しいアカウントを作成しましょう"
    >
      <SignupForm />
    </AuthFormLayout>
  );
}
