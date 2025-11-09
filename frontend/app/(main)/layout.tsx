import { Footer } from "@/features/footer/components/footer";
import { Header } from "@/features/header/components/header";
import { SidebarLeft } from "@/features/sidebar/components/sidebar-left";
import { SidebarRight } from "@/features/sidebar/components/sidebar-right";
import { SidebarProvider } from "@/features/sidebar/contexts/sidebar-context";

export default function MainLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <SidebarProvider>
      <div className="min-h-screen bg-background flex flex-col sm:flex-row">
        <Header />

        <SidebarLeft />

        <main className="flex-1 overflow-y-auto">{children}</main>

        <SidebarRight />

        <Footer />
      </div>
    </SidebarProvider>
  );
}
