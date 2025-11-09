"use client";

import { ProfileMinibar } from "@/features/sidebar/components/profile-minibar";
import { LINK } from "@/lib/links";
import { navItems } from "@/lib/nav-items";
import { cn } from "@/lib/utils";
import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";

export const SidebarLeft = () => {
  const pathname = usePathname();

  return (
    <aside className="hidden w-80 border-r border-border bg-background h-screen sticky top-0 sm:block">
      <div className="h-full flex flex-col p-6">
        <div className="flex-1 overflow-y-auto">
          <Link href={LINK.home} className="flex items-center gap-2 mb-8">
            <Image
              src="/muzee-logo.png"
              alt="Logo"
              width={36}
              height={36}
              className="rounded-lg"
            />
            <span className="font-bold text-lg">Muzee</span>
          </Link>

          <nav className="space-y-3 mb-8">
            {navItems.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                className={cn(
                  "flex items-center gap-3 px-4 py-2.5 rounded-lg text-sm font-medium transition-colors",
                  pathname === item.href
                    ? "bg-primary/10 text-primary"
                    : "text-foreground/70 hover:bg-accent hover:text-foreground",
                )}
              >
                {item.icon}
                <span>{item.label}</span>
              </Link>
            ))}
          </nav>

          <Link
            href="/create"
            className="block w-full bg-primary text-primary-foreground rounded-lg py-2.5 px-4 font-medium text-center hover:bg-primary/90 transition-colors mb-8"
          >
            + 展示を作成
          </Link>
        </div>

        <div className="mt-auto pt-4 border-t border-border">
          <ProfileMinibar />
        </div>
      </div>
    </aside>
  );
};
