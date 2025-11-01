import { AuthFormLayout } from "@/features/auth/components/auth-form-layout";
import { LoginForm } from "@/features/auth/components/login-form";

export default function LoginPage() {
  return (
    <AuthFormLayout
      title="Welcome Back"
      description="アカウントにログインしてください"
    >
      <LoginForm />
    </AuthFormLayout>
  );
}
