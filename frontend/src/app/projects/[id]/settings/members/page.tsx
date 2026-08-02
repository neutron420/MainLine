"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import { useState } from "react";
import { ArrowLeft, Search, Users, UserPlus, Mail, MoreHorizontal, Loader2 } from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbLink,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { NotificationsPopover } from "@/components/notifications-popover";
import {
  useMembers,
  useAddMember,
  useUpdateMemberRole,
  useRemoveMember,
} from "@/lib/api/hooks/use-projects";
import { getApiErrorMessage } from "@/lib/api/errors";

function initialsOf(name: string): string {
  return name
    .split(/[\s@._-]+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? "")
    .join("") || "?";
}

function relativeTime(iso: string | undefined): string {
  if (!iso) return "—";
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return "—";
  const mins = Math.floor((Date.now() - then) / 60_000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins} min ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
}

export default function ProjectMembersPage() {
  const params = useParams();
  const projectId = params.id as string;

  const { data: members, isLoading, error } = useMembers(projectId);
  const addMember = useAddMember(projectId);
  const updateRole = useUpdateMemberRole(projectId);
  const removeMember = useRemoveMember(projectId);

  const [inviteOpen, setInviteOpen] = useState(false);
  const [userId, setUserId] = useState("");
  const [role, setRole] = useState("developer");

  const sendInvite = () => {
    const value = userId.trim();
    if (!value || addMember.isPending) return;
    const isEmail = value.includes("@");
    addMember.mutate(
      isEmail ? { email: value, role } : { userId: value, role },
      {
        onSuccess: () => {
          setInviteOpen(false);
          setUserId("");
          setRole("developer");
        },
      },
    );
  };

  return (
    <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
      <AppSidebar />
      <SidebarInset>
        <header className="sticky top-0 flex h-14 shrink-0 items-center gap-2 border-b bg-background px-4">
          <SidebarTrigger className="-ml-1 size-9" />
          <Separator orientation="vertical" className="mr-2 data-[orientation=vertical]:h-5" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem><BreadcrumbLink href="/projects">Projects</BreadcrumbLink></BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem><BreadcrumbLink href={`/projects/${projectId}`}>SchemaHub</BreadcrumbLink></BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem><BreadcrumbLink href={`/projects/${projectId}/settings`}>Settings</BreadcrumbLink></BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem><BreadcrumbPage>Members</BreadcrumbPage></BreadcrumbItem>
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
        <div className="flex flex-1 flex-col gap-6 p-6 max-w-4xl w-full mx-auto">
          <div className="flex items-start gap-4">
            <Link href={`/projects/${projectId}/settings`}>
              <Button variant="ghost" size="icon" className="size-10 shrink-0 mt-0.5">
                <ArrowLeft className="size-4" />
              </Button>
            </Link>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-3 flex-wrap">
                <h1 className="text-2xl font-semibold tracking-tight flex items-center gap-2">
                  <Users className="size-6" />
                  Project Members
                </h1>
                <Badge variant="outline" className="text-[11px]">{members?.length ?? 0} members</Badge>
              </div>
              <p className="text-sm text-muted-foreground mt-1">Who can view and change schemas in this project</p>
            </div>
            <Dialog open={inviteOpen} onOpenChange={setInviteOpen}>
              <DialogTrigger asChild>
                <Button className="h-10 gap-2 shrink-0">
                  <UserPlus className="size-4" />
                  Add Member
                </Button>
              </DialogTrigger>
              <DialogContent className="max-w-md">
                <DialogHeader>
                  <DialogTitle>Add a member</DialogTitle>
                  <DialogDescription>
                    Invite a registered SchemaHub user by email (or user ID) and
                    assign a role.
                  </DialogDescription>
                </DialogHeader>
                <div className="space-y-4 py-2">
                  <div className="space-y-1.5">
                    <Label>Email or user ID</Label>
                    <div className="relative">
                      <Mail className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground" />
                      <Input
                        type="email"
                        placeholder="teammate@company.com"
                        className="pl-8 font-mono"
                        value={userId}
                        onChange={(e) => setUserId(e.target.value)}
                      />
                    </div>
                  </div>
                  <div className="space-y-1.5">
                    <Label>Role</Label>
                    <Select value={role} onValueChange={setRole}>
                      <SelectTrigger className="h-9">
                        <SelectValue placeholder="Select role" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="admin">Admin</SelectItem>
                        <SelectItem value="developer">Developer</SelectItem>
                        <SelectItem value="viewer">Viewer</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  {addMember.isError && (
                    <p className="text-sm text-red-500">{getApiErrorMessage(addMember.error)}</p>
                  )}
                </div>
                <DialogFooter>
                  <Button onClick={sendInvite} disabled={!userId.trim() || addMember.isPending}>
                    {addMember.isPending ? (
                      <>
                        <Loader2 className="size-4 animate-spin" />
                        Adding…
                      </>
                    ) : (
                      "Add Member"
                    )}
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </div>

          <Card>
            <CardHeader className="border-0 pb-0">
              <CardTitle className="text-base flex items-center gap-2">
                <Users className="size-4" />
                Members
              </CardTitle>
            </CardHeader>
            <CardContent className="pt-4">
              {isLoading ? (
                <p className="text-sm text-muted-foreground py-6 text-center">Loading members...</p>
              ) : error ? (
                <p className="text-sm text-red-500 py-6 text-center">{getApiErrorMessage(error)}</p>
              ) : !members || members.length === 0 ? (
                <p className="text-sm text-muted-foreground py-6 text-center">
                  No members yet. Add your first member to collaborate.
                </p>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-[40%]">Member</TableHead>
                      <TableHead>Role</TableHead>
                      <TableHead>Joined</TableHead>
                      <TableHead className="w-[50px] text-right"></TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {members.map((member) => (
                      <TableRow key={member.userId}>
                        <TableCell>
                          <div className="flex items-center gap-3">
                            <div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-medium">
                              {initialsOf(member.displayName || member.email)}
                            </div>
                            <div className="min-w-0">
                              <p className="text-sm font-medium">{member.displayName || member.email}</p>
                              <p className="text-xs text-muted-foreground truncate">{member.email}</p>
                            </div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline" className="text-[10px] px-1.5 py-0 capitalize">{member.role}</Badge>
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {relativeTime(member.joinedAt)}
                        </TableCell>
                        <TableCell className="text-right">
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="icon" className="size-8">
                                <MoreHorizontal className="size-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              {member.role !== "admin" && (
                                <DropdownMenuItem
                                  onClick={() => updateRole.mutate({ userId: member.userId, role: "admin" })}
                                >
                                  Make admin
                                </DropdownMenuItem>
                              )}
                              {member.role !== "viewer" && (
                                <DropdownMenuItem
                                  onClick={() => updateRole.mutate({ userId: member.userId, role: "viewer" })}
                                >
                                  Make viewer
                                </DropdownMenuItem>
                              )}
                              <DropdownMenuItem
                                className="text-red-500"
                                onClick={() => {
                                  if (confirm(`Remove ${member.email} from this project?`)) {
                                    removeMember.mutate(member.userId);
                                  }
                                }}
                              >
                                Remove
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
