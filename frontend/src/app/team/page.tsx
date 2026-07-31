"use client";

import { useState, useMemo } from "react";
import { UserPlus, Search, MoreHorizontal, Users, UserCheck, ShieldCheck, Hourglass, Mail, RotateCcw, UserX } from "lucide-react";

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

interface Member {
  id: string;
  name: string;
  email: string;
  initials: string;
  role: "admin" | "developer" | "viewer";
  status: "active" | "invited" | "suspended";
  joined: string;
}

const members: Member[] = [
  { id: "m1", name: "R.K Singh", email: "rk@mainline.dev", initials: "RK", role: "admin", status: "active", joined: "Jan 2026" },
  { id: "m2", name: "Alice", email: "alice@mainline.dev", initials: "AL", role: "admin", status: "active", joined: "Jan 2026" },
  { id: "m3", name: "Bob", email: "bob@mainline.dev", initials: "BO", role: "developer", status: "active", joined: "Feb 2026" },
  { id: "m4", name: "Charlie", email: "charlie@mainline.dev", initials: "CH", role: "developer", status: "active", joined: "Feb 2026" },
  { id: "m5", name: "Diana", email: "diana@mainline.dev", initials: "DI", role: "admin", status: "active", joined: "Mar 2026" },
  { id: "m6", name: "Eve", email: "eve@mainline.dev", initials: "EV", role: "viewer", status: "invited", joined: "—" },
  { id: "m7", name: "Frank", email: "frank@mainline.dev", initials: "FR", role: "developer", status: "active", joined: "Apr 2026" },
  { id: "m8", name: "Grace", email: "grace@mainline.dev", initials: "GR", role: "viewer", status: "suspended", joined: "May 2026" },
];

const pendingInvites = [
  { id: "i1", email: "eve@mainline.dev", role: "viewer", invitedBy: "Alice", date: "2 days ago" },
  { id: "i2", email: "harsh@mainline.dev", role: "developer", invitedBy: "Bob", date: "5 days ago" },
];

const roleConfig = {
  admin: { label: "Admin", badge: "default" as const },
  developer: { label: "Developer", badge: "secondary" as const },
  viewer: { label: "Viewer", badge: "outline" as const },
};

const statusConfig = {
  active: { label: "Active", dot: "bg-green-500" },
  invited: { label: "Invited", dot: "bg-yellow-500" },
  suspended: { label: "Suspended", dot: "bg-red-500" },
};

const stats = [
  { title: "Total Members", value: 24, icon: Users },
  { title: "Active", value: 21, icon: UserCheck },
  { title: "Admins", value: 3, icon: ShieldCheck },
  { title: "Pending Invites", value: 2, icon: Hourglass },
];

export default function TeamPage() {
  const [search, setSearch] = useState("");
  const [roleFilter, setRoleFilter] = useState("all");
  const [inviteOpen, setInviteOpen] = useState(false);

  const filtered = useMemo(() => {
    return members.filter((m) => {
      const matchSearch =
        m.name.toLowerCase().includes(search.toLowerCase()) ||
        m.email.toLowerCase().includes(search.toLowerCase());
      const matchRole = roleFilter === "all" || m.role === roleFilter;
      return matchSearch && matchRole;
    });
  }, [search, roleFilter]);

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
              <BreadcrumbItem><BreadcrumbPage>Team</BreadcrumbPage></BreadcrumbItem>
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
          {/* Header + Invite */}
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-semibold tracking-tight">Team</h1>
              <p className="text-sm text-muted-foreground mt-1">Manage team members and roles</p>
            </div>
            <Dialog open={inviteOpen} onOpenChange={setInviteOpen}>
              <DialogTrigger asChild>
                <Tooltip delay={0}>
                  <Button className="h-11 gap-2">
                    <UserPlus />
                    Invite Member
                  </Button>
                  <Tooltip.Content>
                    <p>Invite someone to the team</p>
                  </Tooltip.Content>
                </Tooltip>
              </DialogTrigger>
              <DialogContent className="sm:max-w-[480px]">
                <DialogHeader>
                  <DialogTitle>Invite Member</DialogTitle>
                  <DialogDescription>
                    Send an email invitation to join your team.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-5 py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="email">Email Address</Label>
                    <div className="relative">
                      <Mail className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
                      <Input id="email" type="email" placeholder="teammate@company.com" className="h-11 pl-9" />
                    </div>
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="role">Role</Label>
                    <Select defaultValue="developer">
                      <SelectTrigger id="role" className="h-11">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="admin">Admin — full access</SelectItem>
                        <SelectItem value="developer">Developer — can review and migrate</SelectItem>
                        <SelectItem value="viewer">Viewer — read only</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="message">Message (optional)</Label>
                    <Input id="message" placeholder="Brief note for the invite..." className="h-11" />
                  </div>
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setInviteOpen(false)}>Cancel</Button>
                  <Button className="gap-2" onClick={() => setInviteOpen(false)}>
                    <UserPlus />
                    Send Invite
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </div>

          {/* Stats */}
          <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
            {stats.map((stat) => (
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
                        {roleFilter === "all" ? "Role" : roleConfig[roleFilter as keyof typeof roleConfig].label}
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
                    <TableHead className="w-[45%]">Member</TableHead>
                    <TableHead>Role</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Joined</TableHead>
                    <TableHead className="w-[50px] text-right"></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filtered.map((member) => {
                    const role = roleConfig[member.role];
                    const status = statusConfig[member.status];
                    return (
                      <TableRow key={member.id}>
                        <TableCell>
                          <div className="flex items-center gap-3">
                            <Avatar className="size-9">
                              <AvatarFallback className="text-xs">{member.initials}</AvatarFallback>
                            </Avatar>
                            <div className="min-w-0">
                              <p className="font-medium text-sm truncate">{member.name}</p>
                              <p className="text-xs text-muted-foreground truncate">{member.email}</p>
                            </div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge variant={role.badge} className="text-[10px] px-1.5 py-0">{role.label}</Badge>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <div className={`size-1.5 rounded-full ${status.dot}`} />
                            <span className="text-xs text-muted-foreground">{status.label}</span>
                          </div>
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">{member.joined}</TableCell>
                        <TableCell className="text-right">
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="icon" className="size-8">
                                <MoreHorizontal className="size-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <DropdownMenuLabel>{member.name}</DropdownMenuLabel>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem>View Profile</DropdownMenuItem>
                              <DropdownMenuItem>Change Role</DropdownMenuItem>
                              {member.status !== "suspended" ? (
                                <DropdownMenuItem className="text-destructive"><UserX /> Suspend</DropdownMenuItem>
                              ) : (
                                <DropdownMenuItem>Reactivate</DropdownMenuItem>
                              )}
                            </DropdownMenuContent>
                          </DropdownMenu>
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
                  <p className="text-xs text-muted-foreground mt-1">Try adjusting your search or filters</p>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Pending invites */}
          <Card>
            <CardHeader className="border-0">
              <CardTitle className="text-base flex items-center gap-2">
                <Hourglass className="size-4" />
                Pending Invitations
              </CardTitle>
            </CardHeader>
            <CardContent className="pt-0">
              {pendingInvites.map((invite) => (
                <div key={invite.id} className="flex items-center gap-3 py-3 border-b last:border-b-0">
                  <div className="flex size-9 shrink-0 items-center justify-center rounded-full bg-muted">
                    <Mail className="size-4 text-muted-foreground" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium truncate">{invite.email}</p>
                    <p className="text-xs text-muted-foreground">
                      {roleConfig[invite.role as keyof typeof roleConfig].label} · invited by {invite.invitedBy} · {invite.date}
                    </p>
                  </div>
                  <Button variant="outline" size="sm" className="h-8 gap-1.5">
                    <RotateCcw className="size-3.5" />
                    Resend
                  </Button>
                  <Button variant="ghost" size="sm" className="h-8 text-destructive">
                    Cancel
                  </Button>
                </div>
              ))}
            </CardContent>
          </Card>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
