"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import { useState } from "react";
import { ArrowLeft, Search, Save, Trash2, Users } from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
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
import { Tooltip } from "@heroui/react";
import { Checkbox } from "@/components/ui/checkbox";

export default function ProjectSettingsPage() {
  const params = useParams();
  const projectId = params.id as string;
  const [approvalPolicy, setApprovalPolicy] = useState("one-reviewer");
  const [saved, setSaved] = useState(false);

  const save = () => {
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

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
              <BreadcrumbItem><BreadcrumbLink href="/projects">Projects</BreadcrumbLink></BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem><BreadcrumbLink href={`/projects/${projectId}`}>SchemaHub</BreadcrumbLink></BreadcrumbItem>
              <BreadcrumbSeparator />
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
        <div className="flex flex-1 flex-col gap-6 p-6 max-w-2xl w-full mx-auto">
          <div className="flex items-start gap-4">
            <Link href={`/projects/${projectId}`}>
              <Button variant="ghost" size="icon" className="size-10 shrink-0 mt-0.5">
                <ArrowLeft className="size-4" />
              </Button>
            </Link>
            <div>
              <h1 className="text-2xl font-semibold tracking-tight">Project Settings</h1>
              <p className="text-sm text-muted-foreground mt-1">Manage the SchemaHub project configuration</p>
            </div>
          </div>

          <Card>
            <CardHeader className="border-0">
              <CardTitle className="text-base">General</CardTitle>
            </CardHeader>
            <CardContent className="pt-0 space-y-4">
              <div className="space-y-1.5">
                <Label>Project name</Label>
                <Input defaultValue="SchemaHub" />
              </div>
              <div className="space-y-1.5">
                <Label>Description</Label>
                <Textarea defaultValue="Mainline schema for the core platform — users, teams, billing and analytics." className="min-h-[80px]" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="border-0">
              <CardTitle className="text-base">Migration policy</CardTitle>
            </CardHeader>
            <CardContent className="pt-0 space-y-4">
              <div className="space-y-1.5">
                <Label>Approval policy</Label>
                <Select value={approvalPolicy} onValueChange={setApprovalPolicy}>
                  <SelectTrigger className="h-9">
                    <SelectValue placeholder="Select policy" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="one-reviewer">Single reviewer</SelectItem>
                    <SelectItem value="one-admin">Single admin</SelectItem>
                    <SelectItem value="two-admins">Two admins</SelectItem>
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">
                  {approvalPolicy === "two-admins" ? "Migrations require approval from two different admins." : approvalPolicy === "one-admin" ? "Migrations require approval from any admin." : "Migrations require approval from any reviewer."}
                </p>
              </div>
              <div className="flex items-start gap-2.5 pt-2">
                <Checkbox id="auto-prod" className="mt-0.5" />
                <div>
                  <Label htmlFor="auto-prod">Block migrations to Production without review</Label>
                  <p className="text-xs text-muted-foreground mt-0.5">Reject any migration targeting a Production connection that has not passed review.</p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="border-0">
              <CardTitle className="text-base flex items-center gap-2">
                <Users className="size-4" />
                Members
              </CardTitle>
            </CardHeader>
            <CardContent className="pt-0">
              <p className="text-sm text-muted-foreground">Manage who has access to this project and their roles.</p>
              <Link href={`/projects/${projectId}/settings/members`}>
                <Button variant="outline" size="sm" className="h-9 gap-2 mt-3">
                  <Users className="size-4" />
                  Manage Members
                </Button>
              </Link>
            </CardContent>
          </Card>

          <div className="flex items-center gap-3">
            <Button className="h-9 gap-2" onClick={save}>
              <Save className="size-4" />
              Save Changes
            </Button>
            {saved && <span className="text-sm text-emerald-500">Changes saved</span>}
          </div>

          <Card className="border-red-500/40">
            <CardHeader className="border-0">
              <CardTitle className="text-base text-red-500">Danger zone</CardTitle>
            </CardHeader>
            <CardContent className="pt-0">
              <p className="text-sm text-muted-foreground">
                Deleting this project removes all schemas, migrations and drift history. This cannot be undone.
              </p>
              <Button variant="destructive" size="sm" className="h-9 gap-2 mt-3">
                <Trash2 className="size-4" />
                Delete Project
              </Button>
            </CardContent>
          </Card>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
