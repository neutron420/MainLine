"use client";

import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { useEffect, useState } from "react";
import { ArrowLeft, Search, Save, Trash2, Users, Loader2 } from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
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
  useProject,
  useUpdateProject,
  useDeleteProject,
} from "@/lib/api/hooks/use-projects";
import { getApiErrorMessage } from "@/lib/api/errors";

export default function ProjectSettingsPage() {
  const params = useParams();
  const router = useRouter();
  const projectId = params.id as string;

  const { data: project, isLoading } = useProject(projectId);
  const updateProject = useUpdateProject(projectId);
  const deleteProject = useDeleteProject();

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    if (project) {
      setName(project.name);
      setDescription(project.description ?? "");
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [project?.id]);

  const save = () => {
    updateProject.mutate(
      {
        name: name.trim() || undefined,
        description: description.trim() || undefined,
      },
      {
        onSuccess: () => {
          setSaved(true);
          setTimeout(() => setSaved(false), 2000);
        },
      },
    );
  };

  const removeProject = () => {
    if (!confirm(`Delete project "${project?.name}"? This cannot be undone.`)) return;
    deleteProject.mutate(projectId, {
      onSuccess: () => router.push("/projects"),
    });
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

          {isLoading ? (
            <p className="text-sm text-muted-foreground">Loading project...</p>
          ) : !project ? (
            <p className="text-sm text-red-500">Project not found.</p>
          ) : (
            <>
              <Card>
                <CardHeader className="border-0">
                  <CardTitle className="text-base">General</CardTitle>
                </CardHeader>
                <CardContent className="pt-0 space-y-4">
                  <div className="space-y-1.5">
                    <Label>Project name</Label>
                    <Input value={name} onChange={(e) => setName(e.target.value)} />
                  </div>
                  <div className="space-y-1.5">
                    <Label>Description</Label>
                    <Textarea
                      value={description}
                      onChange={(e) => setDescription(e.target.value)}
                      className="min-h-[80px]"
                    />
                  </div>
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-muted-foreground">Slug</span>
                    <span className="font-mono text-xs">{project.slug}</span>
                  </div>
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-muted-foreground">Visibility</span>
                    <span className="capitalize">{project.visibility}</span>
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
                <Button className="h-9 gap-2" onClick={save} disabled={updateProject.isPending}>
                  {updateProject.isPending ? (
                    <>
                      <Loader2 className="size-4 animate-spin" />
                      Saving…
                    </>
                  ) : (
                    <>
                      <Save className="size-4" />
                      Save Changes
                    </>
                  )}
                </Button>
                {saved && <span className="text-sm text-emerald-500">Changes saved</span>}
                {updateProject.isError && (
                  <span className="text-sm text-red-500">{getApiErrorMessage(updateProject.error)}</span>
                )}
              </div>

              <Card className="border-red-500/40">
                <CardHeader className="border-0">
                  <CardTitle className="text-base text-red-500">Danger zone</CardTitle>
                </CardHeader>
                <CardContent className="pt-0">
                  <p className="text-sm text-muted-foreground">
                    Deleting this project removes all schemas, migrations and drift history. This cannot be undone.
                  </p>
                  <Button
                    variant="destructive"
                    size="sm"
                    className="h-9 gap-2 mt-3"
                    onClick={removeProject}
                    disabled={deleteProject.isPending}
                  >
                    {deleteProject.isPending ? (
                      <>
                        <Loader2 className="size-4 animate-spin" />
                        Deleting…
                      </>
                    ) : (
                      <>
                        <Trash2 className="size-4" />
                        Delete Project
                      </>
                    )}
                  </Button>
                  {deleteProject.isError && (
                    <p className="text-sm text-red-500 mt-3">{getApiErrorMessage(deleteProject.error)}</p>
                  )}
                </CardContent>
              </Card>
            </>
          )}
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
