"use client";

import { useState } from "react";
import { User, Bell, Database, Users, Search, Save, Plug, Trash2 } from "lucide-react";
import Link from "next/link";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
} from "@/components/ui/breadcrumb";
import { NotificationsPopover } from "@/components/notifications-popover";
import { Tooltip } from "@heroui/react";

const settingsTabs = [
  { value: "profile", label: "Profile", icon: User },
  { value: "notifications", label: "Notifications", icon: Bell },
  { value: "connections", label: "Database Connections", icon: Database },
  { value: "team", label: "Team Management", icon: Users },
];

const notifications = [
  { id: "n1", title: "Review notifications", description: "Get notified when someone requests or completes a review" },
  { id: "n2", title: "Migration alerts", description: "Email and in-app alerts when a migration succeeds or fails" },
  { id: "n3", title: "Schema drift detection", description: "Warn immediately when drift is detected in any environment" },
  { id: "n4", title: "Weekly digest", description: "A weekly summary of reviews, migrations, and project activity" },
];

const connections = [
  { id: "c1", name: "Production", host: "ep-remote-01.aws.neon.tech", database: "mainline_prod", status: "connected" },
  { id: "c2", name: "Staging", host: "ep-remote-02.aws.neon.tech", database: "mainline_staging", status: "connected" },
  { id: "c3", name: "Legacy CRM", host: "db.internal.crm.io:5432", database: "crm_legacy", status: "failed" },
];

const roles = [
  { name: "Admin", count: 3, description: "Full access — manage members, approve migrations, and edit settings" },
  { name: "Developer", count: 16, description: "Submit schema changes, review, and run migrations" },
  { name: "Viewer", count: 5, description: "Read-only access to schemas, reviews, and migration history" },
];

export default function SettingsPage() {
  const [notifState, setNotifState] = useState<Record<string, boolean>>({
    n1: true,
    n2: true,
    n3: true,
    n4: false,
  });

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
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem><BreadcrumbPage>Settings</BreadcrumbPage></BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
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
        <div className="flex flex-1 flex-col gap-6 p-6">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Settings</h1>
            <p className="text-sm text-muted-foreground mt-1">Manage your profile, connections, and preferences</p>
          </div>

          <Tabs defaultValue="profile" orientation="vertical" className="flex gap-6 w-full">
            <TabsList variant="line" className="h-fit w-[220px] flex-col items-stretch gap-1">
              {settingsTabs.map((tab) => (
                <TabsTrigger key={tab.value} value={tab.value} className="justify-start gap-2.5 px-3 py-2.5">
                  <tab.icon className="size-4" />
                  {tab.label}
                </TabsTrigger>
              ))}
            </TabsList>

            {/* Profile */}
            <TabsContent value="profile" className="flex-1 min-w-0 space-y-6">
              <Card>
                <CardHeader className="border-0">
                  <CardTitle className="text-base">Profile</CardTitle>
                  <CardDescription>Update your personal information</CardDescription>
                </CardHeader>
                <CardContent className="pt-0 space-y-5">
                  <div className="flex items-center gap-4">
                    <Avatar className="size-16">
                      <AvatarFallback className="text-lg">RK</AvatarFallback>
                    </Avatar>
                    <div className="space-y-2">
                      <Button variant="outline" size="sm" className="h-9">Change Avatar</Button>
                      <p className="text-xs text-muted-foreground">JPG or PNG. Max 2 MB.</p>
                    </div>
                  </div>
                  <Separator />
                  <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                    <div className="grid gap-2">
                      <Label htmlFor="name">Full Name</Label>
                      <Input id="name" defaultValue="R.K Singh" className="h-11" />
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="email">Email</Label>
                      <Input id="email" type="email" defaultValue="rk@mainline.dev" className="h-11" />
                    </div>
                  </div>
                  <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                    <div className="grid gap-2">
                      <Label htmlFor="job">Job Title</Label>
                      <Input id="job" placeholder="e.g. Staff Engineer" className="h-11" />
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="company">Company</Label>
                      <Input id="company" placeholder="e.g. Mainline Inc" className="h-11" />
                    </div>
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="bio">Bio</Label>
                    <Textarea id="bio" placeholder="A short bio for your team..." className="min-h-[100px]" />
                  </div>
                  <div className="flex justify-end">
                    <Button className="h-11 gap-2">
                      <Save />
                      Save Changes
                    </Button>
                  </div>
                </CardContent>
              </Card>
            </TabsContent>

            {/* Notifications */}
            <TabsContent value="notifications" className="flex-1 min-w-0 space-y-6">
              <Card>
                <CardHeader className="border-0">
                  <CardTitle className="text-base">Notification Preferences</CardTitle>
                  <CardDescription>Choose what you want to be notified about</CardDescription>
                </CardHeader>
                <CardContent className="pt-0">
                  <div className="divide-y">
                    {notifications.map((n) => (
                      <div key={n.id} className="flex items-center justify-between gap-4 py-4">
                        <div className="min-w-0">
                          <p className="text-sm font-medium">{n.title}</p>
                          <p className="text-xs text-muted-foreground mt-0.5">{n.description}</p>
                        </div>
                        <Switch
                          checked={notifState[n.id]}
                          onCheckedChange={(checked) => setNotifState((prev) => ({ ...prev, [n.id]: checked }))}
                        />
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            </TabsContent>

            {/* Database Connections */}
            <TabsContent value="connections" className="flex-1 min-w-0 space-y-6">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-lg font-semibold tracking-tight">Database Connections</h2>
                  <p className="text-sm text-muted-foreground mt-0.5">Connections used for drift detection and migrations</p>
                </div>
                <Button className="h-10 gap-2">
                  <Plug />
                  Add Connection
                </Button>
              </div>
              <Card>
                <CardContent className="pt-0">
                  <div className="divide-y">
                    {connections.map((conn) => (
                      <div key={conn.id} className="flex items-center gap-4 py-4">
                        <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-muted">
                          <Database className="size-4 text-muted-foreground" />
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <p className="text-sm font-medium">{conn.name}</p>
                            {conn.status === "connected" ? (
                              <Badge variant="default" className="text-[10px] px-1.5 py-0">Connected</Badge>
                            ) : (
                              <Badge variant="destructive" className="text-[10px] px-1.5 py-0">Failed</Badge>
                            )}
                          </div>
                          <p className="font-mono text-xs text-muted-foreground truncate mt-0.5">{conn.host}</p>
                        </div>
                        <span className="font-mono text-xs text-muted-foreground hidden md:block">{conn.database}</span>
                        <Button variant="outline" size="sm" className="h-8">Edit</Button>
                        <Button variant="ghost" size="icon" className="size-8 text-destructive">
                          <Trash2 className="size-4" />
                        </Button>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            </TabsContent>

            {/* Team Management */}
            <TabsContent value="team" className="flex-1 min-w-0 space-y-6">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-lg font-semibold tracking-tight">Team Management</h2>
                  <p className="text-sm text-muted-foreground mt-0.5">Roles and permissions across your team</p>
                </div>
                <Link href="/team">
                  <Button variant="outline" className="h-10 gap-2">
                    <Users />
                    Manage Team
                  </Button>
                </Link>
              </div>
              <Card>
                <CardContent className="pt-0">
                  <div className="divide-y">
                    {roles.map((role) => (
                      <div key={role.name} className="flex items-center gap-4 py-4">
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <p className="text-sm font-medium">{role.name}</p>
                            <Badge variant="outline" className="text-[10px] px-1.5 py-0">{role.count} members</Badge>
                          </div>
                          <p className="text-xs text-muted-foreground mt-0.5">{role.description}</p>
                        </div>
                        <Badge variant="secondary" className="text-[10px] px-1.5 py-0">
                          {role.name === "Admin" ? "Full" : role.name === "Developer" ? "Read + Write" : "Read Only"}
                        </Badge>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
