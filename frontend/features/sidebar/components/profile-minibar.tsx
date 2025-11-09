"use client";

import { Providers } from "@/app/providers";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { generateStaticLink, LINK } from "@/lib/links";
import { settingsItems } from "@/lib/nav-items";
import { useGetV1MeProfile } from "@/src/api/__generated__/user-profiles/user-profiles";
import { ChevronsUpDown } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

const _ProfileMinibar = () => {
  const router = useRouter();
  const { data: profile } = useGetV1MeProfile();

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
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          className="w-full h-auto flex items-center gap-3 px-3 py-2 rounded-full hover:bg-muted/50 transition-colors justify-start"
        >
          <Avatar className="h-10 w-10">
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

          <div className="flex-1 min-w-0 text-left">
            <p className="font-medium text-sm truncate">
              {profile.user_profile.name}
            </p>
            <p className="text-xs text-muted-foreground truncate">
              @{profile.user_profile.username}
            </p>
          </div>

          <ChevronsUpDown className="h-4 w-4 flex-shrink-0" />
        </Button>
      </DropdownMenuTrigger>

      <DropdownMenuContent
        className="w-[var(--radix-dropdown-menu-trigger-width)]"
        align="start"
        side="top"
        sideOffset={8}
      >
        {settingsItems.map((item) => (
          <DropdownMenuItem key={item.href} asChild>
            <Link href={item.href} className="flex items-center gap-2">
              {item.icon}
              <span>{item.label}</span>
            </Link>
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
};

export const ProfileMinibar = () => {
  return (
    <Providers>
      <_ProfileMinibar />
    </Providers>
  );
};
