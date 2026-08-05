"use client";

import { useEffect, useMemo, useState } from "react";
import { User, Bell, Database, Users, Search, Save, Plug, KeyRound, ExternalLink } from "lucide-react";
import Link from "next/link";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
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
import { ProjectConnectionsCard } from "@/components/project-connections-card";
import {
  useChangePassword,
  useCurrentUser,
  useUpdateProfile,
  formatApiError,
} from "@/lib/api/hooks/use-auth";
import { useMembers, useProjects, type Project } from "@/lib/api/hooks/use-projects";

const settingsTabs = [
  { value: "profile", label: "Profile", icon: User },
  { value: "notifications", label: "Notifications", icon: Bell },
  { value: "connections", label: "Database Connections", icon: Database },
  { value: "team", label: "Team Management", icon: Users },
];

const notifications = [
  { id: "n1", title: "Migration alerts", description: "Email and in-app alerts when a migration succeeds or fails" },
  { id: "n2", title: "Schema drift detection", description: "Warn immediately when drift is detected in any environment" },
  { id: "n3", title: "Member activity", description: "Updates when members join, leave, or change roles" },
  { id: "n4", title: "Weekly digest", description: "A weekly summary of migrations and project activity" },
];

const NOTIF_STORAGE_KEY = "schemahub.notification_prefs";

function loadNotifState(): Record<string, boolean> {
  const defaults: Record<string, boolean> = { n1: true, n2: true, n3: true, n4: false };
  if (typeof window === "undefined") return defaults;
  try {
    const raw = window.localStorage.getItem(NOTIF_STORAGE_KEY);
    if (!raw) return defaults;
    return { ...defaults, ...(JSON.parse(raw) as Record<string, boolean>) };
  } catch {
    return defaults;
  }
}

function saveNotifState(state: Record<string, boolean>) {
  if (typeof window === "undefined") return;
  window.localStorage.setItem(NOTIF_STORAGE_KEY, JSON.stringify(state));
}

function ProjectMemberStats({ project }: { project: Project }) {
  const { data: members, isLoading } = useMembers(project.id);

  const counts = useMemo(() => {
    const c = { admin: 0, developer: 0, viewer: 0 };
    for (const m of members ?? []) {
      if (m.role in c) c[m.role as keyof typeof c] += 1;
    }
    return c;
  }, [members]);

  return (
    <div className="flex items-center justify-between gap-4 rounded-lg border px-4 py-3.5">
      <div className="flex items-center gap-2 min-w-0">
        <p className="text-sm font-medium truncate">{project.name}</p>
        <Badge variant="outline" className="text-[10px] px-1.5 py-0">{project.slug}</Badge>
      </div>
      <div className="flex items-center gap-3 text-xs text-muted-foreground shrink-0">
        {isLoading ? (
          <span>Loading members...</span>
        ) : (
          <>
            <span>{counts.admin} admins</span>
            <span className="text-muted-foreground/40">·</span>
            <span>{counts.developer} developers</span>
            <span className="text-muted-foreground/40">·</span>
            <span>{counts.viewer} viewers</span>
          </>
        )}
      </div>
      <Link href={`/projects/${project.id}/settings/members`} className="text-xs text-primary hover:underline inline-flex items-center gap-1 shrink-0">
        Manage <ExternalLink className="size-3" />
      </Link>
    </div>
  );
}

function TeamRolesSummary() {
  const { data: projects } = useProjects();

  return (
    <div className="flex flex-col gap-3">
      {projects?.map((project) => (
        <ProjectMemberStats key={project.id} project={project} />
      ))}
      {(!projects || projects.length === 0) && (
        <p className="text-sm text-muted-foreground">No projects yet.</p>
      )}
    </div>
  );
}

export default function SettingsPage() {
  const { data: user } = useCurrentUser();
  const updateProfile = useUpdateProfile();
  const changePassword = useChangePassword();
  const { data: projects } = useProjects();

  const [displayName, setDisplayName] = useState("");
  const [saveState, setSaveState] = useState<{ ok?: boolean; message: string } | null>(null);
  const [notifState, setNotifState] = useState<Record<string, boolean>>(loadNotifState);
  const [pw, setPw] = useState({ current: "", next: "", confirm: "" });
  const [pwState, setPwState] = useState<{ ok?: boolean; message: string } | null>(null);

  useEffect(() => {
    if (user?.displayName) setDisplayName(user.displayName);
  }, [user?.displayName]);

  const saveProfile = async () => {
    setSaveState(null);
    try {
      await updateProfile.mutateAsync({ displayName });
      setSaveState({ ok: true, message: "Profile updated" });
    } catch (err) {
      setSaveState({ ok: false, message: formatApiError(err) });
    }
  };

  const submitPassword = async () => {
    setPwState(null);
    if (pw.next.length < 6) {
      setPwState({ ok: false, message: "New password must be at least 6 characters" });
      return;
    }
    if (pw.next !== pw.confirm) {
      setPwState({ ok: false, message: "Passwords do not match" });
      return;
    }
    try {
      await changePassword.mutateAsync({ currentPassword: pw.current, newPassword: pw.next });
      setPwState({ ok: true, message: "Password changed" });
      setPw({ current: "", next: "", confirm: "" });
    } catch (err) {
      setPwState({ ok: false, message: formatApiError(err) });
    }
  };

  const initials = (user?.displayName || "U")
    .split(" ")
    .map((part) => part[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();

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
                      <AvatarFallback className="text-lg">{initials}</AvatarFallback>
                    </Avatar>
                    <div className="space-y-1">
                      <p className="text-sm font-medium">{user?.displayName || "You"}</p>
                      <p className="text-xs text-muted-foreground">{user?.email}</p>
                    </div>
                  </div>
                  <Separator />
                  <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                    <div className="grid gap-2">
                      <Label htmlFor="name">Full Name</Label>
                      <Input
                        id="name"
                        value={displayName}
                        onChange={(e) => setDisplayName(e.target.value)}
                        className="h-11"
                      />
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="email">Email</Label>
                      <Input id="email" type="email" value={user?.email ?? ""} disabled className="h-11" />
                    </div>
                  </div>
                  {saveState && (
                    <p className={`text-sm ${saveState.ok ? "text-emerald-600" : "text-destructive"}`}>
                      {saveState.message}
                    </p>
                  )}
                  <div className="flex justify-end">
                    <Button className="h-11 gap-2" onClick={saveProfile} disabled={updateProfile.isPending}>
                      <Save />
                      {updateProfile.isPending ? "Saving..." : "Save Changes"}
                    </Button>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="border-0">
                  <CardTitle className="text-base flex items-center gap-2">
                    <KeyRound className="size-4" />
                    Change Password
                  </CardTitle>
                </CardHeader>
                <CardContent className="pt-0 space-y-5">
                  <div className="grid gap-2">
                    <Label htmlFor="current-pw">Current Password</Label>
                    <Input
                      id="current-pw"
                      type="password"
                      value={pw.current}
                      onChange={(e) => setPw((p) => ({ ...p, current: e.target.value }))}
                      className="h-11"
                      autoComplete="current-password"
                    />
                  </div>
                  <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                    <div className="grid gap-2">
                      <Label htmlFor="new-pw">New Password</Label>
                      <Input
                        id="new-pw"
                        type="password"
                        value={pw.next}
                        onChange={(e) => setPw((p) => ({ ...p, next: e.target.value }))}
                        className="h-11"
                        autoComplete="new-password"
                      />
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="confirm-pw">Confirm New Password</Label>
                      <Input
                        id="confirm-pw"
                        type="password"
                        value={pw.confirm}
                        onChange={(e) => setPw((p) => ({ ...p, confirm: e.target.value }))}
                        className="h-11"
                        autoComplete="new-password"
                      />
                    </div>
                  </div>
                  {pwState && (
                    <p className={`text-sm ${pwState.ok ? "text-emerald-600" : "text-destructive"}`}>
                      {pwState.message}
                    </p>
                  )}
                  <div className="flex justify-end">
                    <Button className="h-11 gap-2" onClick={submitPassword} disabled={changePassword.isPending}>
                      {changePassword.isPending ? "Updating..." : "Update Password"}
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
                  <CardDescription>Choose what you want to be notified about. Saved locally in this browser.</CardDescription>
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
                          onCheckedChange={(checked) =>
                            setNotifState((prev) => {
                              const next = { ...prev, [n.id]: checked };
                              saveNotifState(next);
                              return next;
                            })
                          }
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
                <Link href="/projects">
                  <Button className="h-10 gap-2">
                    <Plug />
                    Add Connection
                  </Button>
                </Link>
              </div>
              <Card>
                <CardContent className="pt-0">
                  <div className="space-y-6">
                    {projects?.map((project) => (
                      <ProjectConnectionsCard key={project.id} project={project} />
                    ))}
                    {(!projects || projects.length === 0) && (
                      <p className="text-sm text-muted-foreground">No projects yet.</p>
                    )}
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
              <TeamRolesSummary />
            </TabsContent>
          </Tabs>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
