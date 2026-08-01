"use client";

import Link from "next/link";
import { useState } from "react";
import { Search, Github, Globe, MessageSquare, Link2, Database } from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { NotificationsPopover } from "@/components/notifications-popover";
import { Tooltip } from "@heroui/react";
import { dbConnections, connectionStatusConfig } from "@/lib/connections-data";

type AccountState = "connected" | "disconnected";

const accounts = [
  { id: "github", name: "GitHub", description: "Sign in and link repositories", icon: Github, state: "connected" as AccountState, email: "aarav@schemahub.dev" },
  { id: "google", name: "Google", description: "Sign in with your Google account", icon: Globe, state: "connected" as AccountState, email: "aarav@gmail.com" },
  { id: "slack", name: "Slack", description: "Get schema alerts in Slack", icon: MessageSquare, state: "disconnected" as AccountState, email: "" },
];

export default function LinkedAccountsPage() {
  const [linked, setLinked] = useState<Record<string, AccountState>>(
    Object.fromEntries(accounts.map((a) => [a.id, a.state]))
  );

  return (
    <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
      <AppSidebar />
      <SidebarInset>
        <header className="sticky top-0 flex h-14 shrink-0 items-center gap-2 border-b bg-background px-4">
          <Tooltip delay={0}>
            <SidebarTrigger className="-ml-1 size-9" />
            <Tooltip.Content>
              <p>Toggle sidebar</p>
            </Tooltip.Content>
          </Tooltip>
          <Separator orientation="vertical" className="mr-2 data-[orientation=vertical]:h-5" />
          <div className="flex items-center gap-2 ml-auto">
            <div className="relative">
              <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground" />
              <Input
                placeholder="Search..."
                className="w-[180px] lg:w-[220px] h-9 pl-8 text-sm"
              />
            </div>
            <NotificationsPopover />
          </div>
        </header>
        <div className="flex flex-1 flex-col gap-6 p-6 max-w-2xl w-full mx-auto">
          <div>
            <div className="flex items-center gap-3 flex-wrap">
              <h1 className="text-2xl font-semibold tracking-tight flex items-center gap-2">
                <Link2 className="size-6" />
                Linked Accounts
              </h1>
            </div>
            <p className="text-sm text-muted-foreground mt-1">Connect accounts and manage database connections</p>
          </div>

          <Card>
            <CardHeader className="border-0">
              <CardTitle className="text-base">Account connections</CardTitle>
            </CardHeader>
            <CardContent className="pt-0">
              <div className="divide-y">
                {accounts.map((account) => {
                  const isLinked = linked[account.id] === "connected";
                  return (
                    <div key={account.id} className="flex items-center gap-4 py-4">
                      <div className="flex size-10 shrink-0 items-center justify-center rounded-lg bg-muted">
                        <account.icon className="size-5 text-muted-foreground" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium">{account.name}</p>
                        <p className="text-xs text-muted-foreground truncate mt-0.5">
                          {isLinked ? account.email : account.description}
                        </p>
                      </div>
                      {isLinked ? (
                        <Badge variant="default" className="text-[10px] px-1.5 py-0 gap-1">
                          <span className="size-1.5 rounded-full bg-emerald-500" />
                          Connected
                        </Badge>
                      ) : (
                        <Button variant="outline" size="sm" className="h-8 text-xs">
                          Connect
                        </Button>
                      )}
                      <Switch
                        checked={isLinked}
                        onCheckedChange={(v) => setLinked((prev) => ({ ...prev, [account.id]: v ? "connected" : "disconnected" }))}
                      />
                    </div>
                  );
                })}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="border-0">
              <CardTitle className="text-base flex items-center gap-2">
                <Database className="size-4" />
                Connected databases
              </CardTitle>
            </CardHeader>
            <CardContent className="pt-0">
              <div className="divide-y">
                {dbConnections.map((conn) => {
                  const status = connectionStatusConfig[conn.status];
                  return (
                    <div key={conn.id} className="flex items-center gap-4 py-3.5">
                      <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-muted">
                        <Database className="size-4 text-muted-foreground" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium">{conn.name}</p>
                        <p className="text-xs text-muted-foreground truncate font-mono mt-0.5">{conn.host}</p>
                      </div>
                      <Badge variant={status.badge} className="text-[10px] px-1.5 py-0 gap-1">
                        <span className={`size-1.5 rounded-full ${status.dot}`} />
                        {status.label}
                      </Badge>
                    </div>
                  );
                })}
              </div>
              <Separator className="my-4" />
              <Link href="/projects/prj-schemahub/connections/new">
                <Button variant="outline" size="sm" className="h-9">
                  Add Connection
                </Button>
              </Link>
            </CardContent>
          </Card>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
