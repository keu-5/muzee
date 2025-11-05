"use client";

import { navItems, settingsItems } from "@/lib/nav-items";
import { cn } from "@/lib/utils";
import Link from "next/link";
import { usePathname } from "next/navigation";

export const SidebarLeft = () => {
  const pathname = usePathname();

  return (
    <aside className="hidden w-80 border-r border-border bg-background h-screen sticky top-0 overflow-y-auto sm:block">
      <div className="p-6">
        {/* Logo */}
        <Link href="/" className="flex items-center gap-2 mb-8">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-primary to-emerald-600 flex items-center justify-center">
            <span className="font-bold text-white text-lg">M</span>
          </div>
          <span className="font-bold text-lg">Muzee</span>
        </Link>

        {/* Main Navigation */}
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

        {/* Create Exhibition Button */}
        <Link
          href="/create"
          className="w-full bg-primary text-primary-foreground rounded-lg py-2.5 px-4 font-medium text-center hover:bg-primary/90 transition-colors mb-8"
        >
          + 展示を作成
        </Link>

        {/* Divider */}
        <div className="border-t border-border my-6" />

        {/* Settings */}
        <nav className="space-y-3">
          {settingsItems.map((item) => (
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
      </div>
    </aside>
  );
};
