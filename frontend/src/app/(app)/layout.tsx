"use client";

import type { CSSProperties, ReactNode } from "react";
import { usePathname } from "next/navigation";

import { AppSidebar } from "@/components/app-sidebar";
import { NotificationsPopover } from "@/components/notifications-popover";
import { Separator } from "@/components/ui/separator";
import { SidebarInset, SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";

function sectionTitle(pathname: string): string {
  if (pathname.startsWith("/projects")) return "Projects";
  if (pathname.startsWith("/schemas")) return "Schemas";
  if (pathname.startsWith("/team")) return "Team";
  if (pathname.startsWith("/settings")) return "Settings";
  return "Overview";
}

export default function AppLayout({ children }: Readonly<{ children: ReactNode }>) {
  const pathname = usePathname();

  return (
    <SidebarProvider style={{ "--sidebar-width": "350px" } as CSSProperties}>
      <AppSidebar />
      <SidebarInset>
        <header className="sticky top-0 flex h-14 shrink-0 items-center gap-2 border-b bg-background px-4">
          <SidebarTrigger className="-ml-1 size-9" />
          <Separator orientation="vertical" className="mr-2 data-[orientation=vertical]:h-5" />
          <span className="text-sm font-medium text-muted-foreground">{sectionTitle(pathname)}</span>
          <div className="flex items-center gap-2 ml-auto">
            <NotificationsPopover />
          </div>
        </header>
        {children}
      </SidebarInset>
    </SidebarProvider>
  );
}
