import { LogoutButton } from "@/features/auth/components/logout-button";
import { Tests } from "@/features/tests/components/Tests";
import { Profile } from "@/features/user/components/profile";

export default function Home() {
  return (
    <>
      <Profile />
      <LogoutButton />
      <Tests />
    </>
  );
}
