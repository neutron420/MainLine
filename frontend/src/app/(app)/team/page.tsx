"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { UserPlus, Search, Users, UserCheck, ShieldCheck, MoreHorizontal, Hourglass } from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  DropdownMenuCheckboxItem,
} from "@/components/ui/dropdown-menu";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
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
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { NotificationsPopover } from "@/components/notifications-popover";
import { Tooltip } from "@heroui/react";
import {
  useAddMember,
  useMembers,
  useProjects,
  useRemoveMember,
  useUpdateMemberRole,
} from "@/lib/api/hooks/use-projects";
import type { ProjectMember } from "@/lib/gen/project/v1/project_messages_pb";

const roleConfig: Record<string, { label: string; badge: "default" | "secondary" | "outline"; rank: number }> = {
  admin: { label: "Admin", badge: "default", rank: 3 },
  developer: { label: "Developer", badge: "secondary", rank: 2 },
  viewer: { label: "Viewer", badge: "outline", rank: 1 },
};

interface AggregatedMember {
  userId: string;
  displayName: string;
  email: string;
  role: string;
  joinedAt: string;
  projects: string[];
}

function MemberReporter({
  projectId,
  onLoad,
}: {
  projectId: string;
  onLoad: (projectId: string, members: ProjectMember[]) => void;
}) {
  const { data: members } = useMembers(projectId);

  useEffect(() => {
    if (members) onLoad(projectId, members);
  }, [members, projectId, onLoad]);

  return null;
}

function MemberActions({
  member,
  projectId,
}: {
  member: AggregatedMember;
  projectId: string;
}) {
  const updateRole = useUpdateMemberRole(projectId);
  const remove = useRemoveMember(projectId);

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" className="size-8">
          <MoreHorizontal className="size-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuLabel>{member.displayName || member.email}</DropdownMenuLabel>
        <DropdownMenuSeparator />
        {(["admin", "developer", "viewer"] as const).map((role) => (
          <DropdownMenuItem
            key={role}
            disabled={role === member.role}
            onClick={() => updateRole.mutate({ userId: member.userId, role })}
          >
            Make {roleConfig[role].label}
          </DropdownMenuItem>
        ))}
        <DropdownMenuSeparator />
        <DropdownMenuItem
          className="text-destructive"
          disabled={roleConfig[member.role]?.rank >= 3}
          onClick={() => remove.mutate(member.userId)}
        >
          Remove from project
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function AddMemberTrigger({
  projectId,
  userId,
  role,
  trigger,
  onDone,
  onError,
}: {
  projectId: string;
  userId: string;
  role: string;
  trigger: number;
  onDone: () => void;
  onError: (message: string) => void;
}) {
  const add = useAddMember(projectId);

  useEffect(() => {
    if (trigger === 0) return;
    add.mutate(
      { userId, role },
      {
        onSuccess: () => onDone(),
        onError: (err) => onError(err.message),
      }
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [trigger]);

  return null;
}

export default function TeamPage() {
  const { data: projects } = useProjects();
  const [byProject, setByProject] = useState<Record<string, ProjectMember[]>>({});
  const [search, setSearch] = useState("");
  const [roleFilter, setRoleFilter] = useState("all");
  const [inviteOpen, setInviteOpen] = useState(false);
  const [invite, setInvite] = useState({ projectId: "", userId: "", role: "developer" });
  const [inviteError, setInviteError] = useState<string | null>(null);
  const [inviteTrigger, setInviteTrigger] = useState(0);

  const onMembersLoaded = useCallback((projectId: string, members: ProjectMember[]) => {
    setByProject((prev) => {
      if (prev[projectId] === members) return prev;
      return { ...prev, [projectId]: members };
    });
  }, []);

  const aggregated = useMemo(() => {
    const map = new Map<string, AggregatedMember>();
    for (const [projectId, members] of Object.entries(byProject)) {
      for (const m of members) {
        const existing = map.get(m.userId);
        if (existing) {
          if (!existing.projects.includes(projectId)) existing.projects.push(projectId);
          const cur = roleConfig[existing.role]?.rank ?? 0;
          const next = roleConfig[m.role]?.rank ?? 0;
          if (next > cur) {
            existing.role = m.role;
            existing.joinedAt = m.joinedAt;
          }
        } else {
          map.set(m.userId, {
            userId: m.userId,
            displayName: m.displayName,
            email: m.email,
            role: m.role,
            joinedAt: m.joinedAt,
            projects: [projectId],
          });
        }
      }
    }
    return Array.from(map.values());
  }, [byProject]);

  const stats = useMemo(() => {
    return {
      total: aggregated.length,
      active: aggregated.filter((m) => m.projects.length > 0).length,
      admins: aggregated.filter((m) => m.role === "admin").length,
      devs: aggregated.filter((m) => m.role === "developer").length,
      viewers: aggregated.filter((m) => m.role === "viewer").length,
    };
  }, [aggregated]);

  const filtered = useMemo(() => {
    return aggregated.filter((m) => {
      const matchSearch =
        m.displayName.toLowerCase().includes(search.toLowerCase()) ||
        m.email.toLowerCase().includes(search.toLowerCase());
      const matchRole = roleFilter === "all" || m.role === roleFilter;
      return matchSearch && matchRole;
    });
  }, [aggregated, search, roleFilter]);

  const submitInvite = () => {
    setInviteError(null);
    if (!invite.projectId) {
      setInviteError("Select a project");
      return;
    }
    if (!invite.userId.trim()) {
      setInviteError("Enter the user ID to add");
      return;
    }
    setInviteTrigger((t) => t + 1);
  };

  return (
            <div className="flex flex-1 flex-col gap-6 p-6">
        <div className="flex flex-wrap items-center gap-3">
            <div className="relative">
              <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground" />
              <Input
                placeholder="Search..."
                className="w-[180px] lg:w-[220px] h-9 pl-8 text-sm"
              />
            </div>
          </div>
          {projects?.map((project) => (
            <MemberReporter key={project.id} projectId={project.id} onLoad={onMembersLoaded} />
          ))}

          {/* Header + Invite */}
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-semibold tracking-tight">Team</h1>
              <p className="text-sm text-muted-foreground mt-1">Members across all your projects</p>
            </div>
            <Dialog open={inviteOpen} onOpenChange={setInviteOpen}>
              <DialogTrigger asChild>
                <Tooltip delay={0}>
                  <Button className="h-11 gap-2">
                    <UserPlus />
                    Add Member
                  </Button>
                  <Tooltip.Content>
                    <p>Add a member to a project</p>
                  </Tooltip.Content>
                </Tooltip>
              </DialogTrigger>
              <DialogContent className="sm:max-w-[480px]">
                <DialogHeader>
                  <DialogTitle>Add Member</DialogTitle>
                  <DialogDescription>
                    Add an existing user to a project with a role.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-5 py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="invite-project">Project</Label>
                    <Select value={invite.projectId} onValueChange={(v) => setInvite((p) => ({ ...p, projectId: v }))}>
                      <SelectTrigger id="invite-project" className="h-11">
                        <SelectValue placeholder="Select project" />
                      </SelectTrigger>
                      <SelectContent>
                        {projects?.map((p) => (
                          <SelectItem key={p.id} value={p.id}>{p.name}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="invite-user">User ID</Label>
                    <Input
                      id="invite-user"
                      placeholder="Enter the user's ID"
                      value={invite.userId}
                      onChange={(e) => setInvite((p) => ({ ...p, userId: e.target.value }))}
                      className="h-11"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="invite-role">Role</Label>
                    <Select value={invite.role} onValueChange={(v) => setInvite((p) => ({ ...p, role: v }))}>
                      <SelectTrigger id="invite-role" className="h-11">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="admin">Admin — full access</SelectItem>
                        <SelectItem value="developer">Developer — can review and migrate</SelectItem>
                        <SelectItem value="viewer">Viewer — read only</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  {inviteError && <p className="text-destructive text-sm">{inviteError}</p>}
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setInviteOpen(false)}>Cancel</Button>
                  <Button className="gap-2" onClick={submitInvite}>
                    <UserPlus />
                    Add Member
                  </Button>
                </DialogFooter>
                {invite.projectId && (
                  <AddMemberTrigger
                    projectId={invite.projectId}
                    userId={invite.userId.trim()}
                    role={invite.role}
                    trigger={inviteTrigger}
                    onDone={() => {
                      setInviteOpen(false);
                      setInvite({ projectId: "", userId: "", role: "developer" });
                      setInviteTrigger(0);
                    }}
                    onError={(message) => setInviteError(message)}
                  />
                )}
              </DialogContent>
            </Dialog>
          </div>

          {/* Stats */}
          <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
            {[
              { title: "Total Members", value: stats.total, icon: Users },
              { title: "Admins", value: stats.admins, icon: ShieldCheck },
              { title: "Developers", value: stats.devs, icon: UserCheck },
              { title: "Viewers", value: stats.viewers, icon: Hourglass },
            ].map((stat) => (
              <Card key={stat.title}>
                <CardContent className="flex items-center gap-4 p-5">
                  <div className="flex size-11 shrink-0 items-center justify-center rounded-lg bg-muted">
                    <stat.icon className="size-5 text-muted-foreground" />
                  </div>
                  <div>
                    <p className="text-2xl font-semibold tracking-tight">{stat.value}</p>
                    <p className="text-xs text-muted-foreground">{stat.title}</p>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          {/* Members */}
          <Card>
            <CardHeader className="border-0">
              <div className="flex items-center justify-between gap-4">
                <CardTitle className="text-base">Members</CardTitle>
                <div className="flex items-center gap-3">
                  <div className="relative">
                    <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground" />
                    <Input
                      placeholder="Search members..."
                      value={search}
                      onChange={(e) => setSearch(e.target.value)}
                      className="w-[220px] h-9 pl-8 text-sm"
                    />
                  </div>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="outline" className="h-9">
                        {roleFilter === "all" ? "Role" : roleConfig[roleFilter].label}
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent>
                      <DropdownMenuLabel>Filter by Role</DropdownMenuLabel>
                      <DropdownMenuSeparator />
                      <DropdownMenuCheckboxItem checked={roleFilter === "all"} onCheckedChange={() => setRoleFilter("all")}>All Roles</DropdownMenuCheckboxItem>
                      <DropdownMenuCheckboxItem checked={roleFilter === "admin"} onCheckedChange={() => setRoleFilter("admin")}>Admin</DropdownMenuCheckboxItem>
                      <DropdownMenuCheckboxItem checked={roleFilter === "developer"} onCheckedChange={() => setRoleFilter("developer")}>Developer</DropdownMenuCheckboxItem>
                      <DropdownMenuCheckboxItem checked={roleFilter === "viewer"} onCheckedChange={() => setRoleFilter("viewer")}>Viewer</DropdownMenuCheckboxItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </div>
            </CardHeader>
            <CardContent className="pt-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[40%]">Member</TableHead>
                    <TableHead>Role</TableHead>
                    <TableHead>Projects</TableHead>
                    <TableHead>Joined</TableHead>
                    <TableHead className="w-[50px] text-right"></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filtered.map((member) => {
                    const role = roleConfig[member.role] ?? roleConfig.viewer;
                    const initials = (member.displayName || member.email || "?")
                      .split(/[\s@]/)
                      .filter(Boolean)
                      .map((part) => part[0])
                      .join("")
                      .slice(0, 2)
                      .toUpperCase();
                    return (
                      <TableRow key={member.userId}>
                        <TableCell>
                          <div className="flex items-center gap-3">
                            <Avatar className="size-9">
                              <AvatarFallback className="text-xs">{initials}</AvatarFallback>
                            </Avatar>
                            <div className="min-w-0">
                              <p className="font-medium text-sm truncate">{member.displayName}</p>
                              <p className="text-xs text-muted-foreground truncate">{member.email}</p>
                            </div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge variant={role.badge} className="text-[10px] px-1.5 py-0">{role.label}</Badge>
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {member.projects.length}
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {member.joinedAt ? new Date(member.joinedAt).toLocaleDateString() : "—"}
                        </TableCell>
                        <TableCell className="text-right">
                          {member.projects[0] && (
                            <MemberActions member={member} projectId={member.projects[0]} />
                          )}
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
              {filtered.length === 0 && (
                <div className="flex flex-col items-center justify-center py-12 text-center">
                  <Users className="size-10 text-muted-foreground/40 mb-3" />
                  <h3 className="text-sm font-medium">No members found</h3>
                  <p className="text-xs text-muted-foreground mt-1">
                    {aggregated.length === 0 ? "No projects or members yet" : "Try adjusting your search or filters"}
                  </p>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
  );
}
