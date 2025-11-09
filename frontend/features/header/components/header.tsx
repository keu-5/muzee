"use client";

import { Providers } from "@/app/providers";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { useSidebar } from "@/features/sidebar/contexts/sidebar-context";
import { generateStaticLink, LINK } from "@/lib/links";
import { useGetV1MeProfile } from "@/src/api/__generated__/user-profiles/user-profiles";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

const _Header = () => {
  const router = useRouter();
  const { data: profile } = useGetV1MeProfile();
  const { open } = useSidebar();

  useEffect(() => {
    if (!profile?.user_profile) {
      router.push(LINK.createProfile);
    }
  }, [profile, router]);

  if (!profile?.user_profile) {
    return null;
  }

  const iconUrl = generateStaticLink(profile.user_profile.icon_path);

  return (
    <header className="w-full h-16 border-b border-border bg-background sticky top-0 z-10 sm:hidden flex items-center justify-between px-4">
      <Button
        variant="ghost"
        size="icon"
        className="h-9 w-9 rounded-full p-0"
        onClick={open}
        aria-label="メニューを開く"
      >
        <Avatar className="h-9 w-9">
          {iconUrl ? (
            <>
              <AvatarImage src={iconUrl} alt="User Avatar" />
              <AvatarFallback>U</AvatarFallback>
            </>
          ) : (
            <AvatarFallback>U</AvatarFallback>
          )}
          <AvatarFallback>{profile.user_profile.username}</AvatarFallback>
        </Avatar>
      </Button>

      <div className="absolute left-1/2 -translate-x-1/2">
        <p className="text-lg font-semibold tracking-wide">Muzee</p>
      </div>

      <div className="w-9" />
    </header>
  );
};

export const Header = () => {
  return (
    <Providers>
      <_Header />
    </Providers>
  );
};
