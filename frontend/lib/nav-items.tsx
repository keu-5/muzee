import { Compass, Heart, Home, LogOut, Settings, User } from "lucide-react";

interface NavItem {
  icon: React.ReactNode;
  label: string;
  href: string;
}

export const navItems: NavItem[] = [
  { icon: <Home className="w-5 h-5" />, label: "ホーム", href: "/home" },
  {
    icon: <Compass className="w-5 h-5" />,
    label: "おすすめ",
    href: "/recommendations",
  },
  {
    icon: <Heart className="w-5 h-5" />,
    label: "お気に入り",
    href: "/favorites",
  },
  {
    icon: <User className="w-5 h-5" />,
    label: "プロフィール",
    href: "/profile/user",
  },
];

export const settingsItems: NavItem[] = [
  { icon: <Settings className="w-5 h-5" />, label: "設定", href: "/settings" },
  {
    icon: <LogOut className="w-5 h-5" />,
    label: "ログアウト",
    href: "/logout",
  },
];
