"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { ArrowLeft, FolderGit2, Plus, Loader2, CheckCircle2 } from "lucide-react";

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
import { Badge } from "@/components/ui/badge";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { NotificationsPopover } from "@/components/notifications-popover";
import { Tooltip } from "@heroui/react";
import { useCreateProject } from "@/lib/api/hooks/use-projects";
import { getApiErrorMessage } from "@/lib/api/errors";

const templates = [
  {
    id: "blank",
    name: "Blank",
    description: "Start with a single public schema and no tables",
    icon: "▦",
  },
  {
    id: "starter",
    name: "Starter",
    description: "Users, teams, sessions — common auth-ready tables",
    icon: "⚡",
  },
  {
    id: "ecommerce",
    name: "E-commerce",
    description: "Products, orders, payments and inventory tables",
    icon: "🛒",
  },
];

export default function CreateProjectPage() {
  const router = useRouter();
  const createProject = useCreateProject();
  const [template, setTemplate] = useState("starter");
  const [name, setName] = useState("SchemaHub");
  const [description, setDescription] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [createdId, setCreatedId] = useState<string | null>(null);

  const create = async () => {
    if (createProject.isPending) return;
    setError(null);
    try {
      const project = await createProject.mutateAsync({
        name: name.trim() || "Untitled project",
        description: description.trim(),
        visibility: "private",
        template,
      });
      setCreatedId(project.id);
      setTimeout(() => router.push(`/projects/${project.id}`), 900);
    } catch (err) {
      setError(getApiErrorMessage(err));
    }
  };

  if (createdId) {
    return (
      <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
        <AppSidebar />
        <SidebarInset>
          <div className="flex flex-1 flex-col items-center justify-center gap-4 p-8">
            <div className="flex size-14 items-center justify-center rounded-full bg-emerald-500/10">
              <CheckCircle2 className="size-7 text-emerald-500" />
            </div>
            <h2 className="text-xl font-semibold">Project created</h2>
            <p className="text-sm text-muted-foreground max-w-sm text-center">
              Redirecting to your project…
            </p>
          </div>
        </SidebarInset>
      </SidebarProvider>
    );
  }

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
            <NotificationsPopover />
          </div>
        </header>
        <div className="flex flex-1 flex-col gap-6 p-6 max-w-2xl w-full mx-auto">
          <div className="flex items-start gap-4">
            <Button variant="ghost" size="icon" className="size-10 shrink-0 mt-0.5" onClick={() => router.push("/projects")}>
              <ArrowLeft className="size-4" />
            </Button>
            <div>
              <div className="flex items-center gap-3 flex-wrap">
                <h1 className="text-2xl font-semibold tracking-tight flex items-center gap-2">
                  <FolderGit2 className="size-6" />
                  New Project
                </h1>
              </div>
              <p className="text-sm text-muted-foreground mt-1">Create a project to start tracking database schemas</p>
            </div>
          </div>

          <Card>
            <CardHeader className="border-0">
              <CardTitle className="text-base">Project details</CardTitle>
            </CardHeader>
            <CardContent className="pt-0 space-y-4">
              <div className="space-y-1.5">
                <Label>Project name</Label>
                <Input
                  placeholder="e.g. SchemaHub"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                />
              </div>
              <div className="space-y-1.5">
                <Label>Description</Label>
                <Textarea
                  placeholder="What is this database for?"
                  className="min-h-[70px]"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                />
              </div>
              {error && (
                <p className="text-destructive text-sm">{error}</p>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="border-0">
              <CardTitle className="text-base">Template</CardTitle>
            </CardHeader>
            <CardContent className="pt-0">
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                {templates.map((t) => (
                  <button
                    key={t.id}
                    type="button"
                    onClick={() => setTemplate(t.id)}
                    className={`flex flex-col items-start gap-2 rounded-lg border p-4 text-left transition-colors ${
                      template === t.id
                        ? "border-primary bg-primary/5 ring-1 ring-primary"
                        : "hover:bg-muted/50"
                    }`}
                  >
                    <span className="text-lg">{t.icon}</span>
                    <div>
                      <p className="text-sm font-medium">{t.name}</p>
                      <p className="text-xs text-muted-foreground mt-0.5 leading-relaxed">{t.description}</p>
                    </div>
                  </button>
                ))}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="border-0">
              <CardTitle className="text-base">Database source</CardTitle>
            </CardHeader>
            <CardContent className="pt-0 space-y-4">
              <div className="space-y-1.5">
                <Label>Source</Label>
                <Select defaultValue="neon">
                  <SelectTrigger className="h-9">
                    <SelectValue placeholder="Select source" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="neon">Neon (managed PostgreSQL)</SelectItem>
                    <SelectItem value="connection">Existing connection</SelectItem>
                    <SelectItem value="none">None — connect later</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="flex items-center justify-between rounded-lg border bg-muted/40 p-3.5">
                <div>
                  <p className="text-sm font-medium">Neon integration</p>
                  <p className="text-xs text-muted-foreground mt-0.5">Create a branch or connect to an existing database</p>
                </div>
                <Badge variant="outline" className="text-[10px] px-1.5 py-0">Available</Badge>
              </div>
            </CardContent>
          </Card>

          <div className="flex items-center gap-3">
            <Button variant="ghost" className="h-10" onClick={() => router.push("/projects")}>
              Cancel
            </Button>
            <Button className="h-10 gap-2 ml-auto" onClick={create} disabled={createProject.isPending}>
              {createProject.isPending ? (
                <>
                  <Loader2 className="size-4 animate-spin" />
                  Creating…
                </>
              ) : (
                <>
                  <Plus className="size-4" />
                  Create Project
                </>
              )}
            </Button>
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
