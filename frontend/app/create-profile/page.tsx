import { ProfileForm } from "@/features/user/components/profile-form";

export default function Page() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-[#BDB76B] via-background to-background relative overflow-hidden p-4">
      <div className="absolute top-20 left-10 w-72 h-72 bg-[#BDB76B] opacity-10 rounded-full blur-3xl"></div>
      <div className="absolute bottom-10 right-20 w-96 h-96 bg-[#BDB76B] opacity-5 rounded-full blur-3xl"></div>

      <div className="relative flex items-center justify-center min-h-screen">
        <img
          src="/modern-abstract-logo.jpg"
          alt="Logo"
          className="absolute top-8 left-8 h-10 w-10 object-cover rounded-lg opacity-80"
        />
        <ProfileForm />
      </div>
    </div>
  );
}
