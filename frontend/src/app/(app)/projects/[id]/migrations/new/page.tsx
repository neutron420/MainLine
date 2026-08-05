"use client";

import { useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, FileCode2, Send, Loader2, CheckCircle2 } from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
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
import { useCreateMigration } from "@/lib/api/hooks/use-migrations";
import { getApiErrorMessage } from "@/lib/api/errors";

export default function CreateMigrationPage() {
  const params = useParams();
  const router = useRouter();
  const projectId = params.id as string;

  const createMigration = useCreateMigration();
  const createdId = createMigration.data?.id;

  const [form, setForm] = useState({
    title: "",
    version: "",
    description: "",
    upSql: "",
    downSql: "",
  });

  const set = (key: keyof typeof form) => (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => setForm((prev) => ({ ...prev, [key]: e.target.value }));

  const canSubmit = form.title.trim() && form.version.trim() && form.upSql.trim();

  const submit = () => {
    if (!canSubmit || createMigration.isPending) return;
    createMigration.mutate(
      {
        projectId,
        title: form.title.trim(),
        version: form.version.trim(),
        upSql: form.upSql,
        downSql: form.downSql.trim() || undefined,
        description: form.description.trim() || undefined,
      },
      {
        onSuccess: (migration) => {
          if (migration?.id) {
            setTimeout(() => router.push(`/projects/${projectId}/migrations/${migration.id}`), 900);
          }
        },
      },
    );
  };

  if (createdId) {
    return (
                <div className="flex flex-1 flex-col items-center justify-center gap-4 p-8">
            <div className="flex size-14 items-center justify-center rounded-full bg-emerald-500/10">
              <CheckCircle2 className="size-7 text-emerald-500" />
            </div>
            <h2 className="text-xl font-semibold">Migration created</h2>
            <p className="text-sm text-muted-foreground max-w-sm text-center">
              <span className="font-mono text-foreground">{form.version}</span> saved as a draft.
            </p>
            <div className="flex items-center gap-3">
              <Link href={`/projects/${projectId}`}>
                <Button variant="outline">Back to Project</Button>
              </Link>
              <Link href={`/projects/${projectId}/migrations/${createdId}`}>
                <Button>View Migration</Button>
              </Link>
            </div>
          </div>
    );
  }

  return (
    <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
      <AppSidebar />
      <SidebarInset>
        <div className="flex flex-1 flex-col gap-6 p-6">
        <div className="flex flex-wrap items-center gap-3">
          </div>
          {/* Header */}
          <div className="flex items-start gap-4">
            <Link href={`/projects/${projectId}`}>
              <Button variant="ghost" size="icon" className="size-10 shrink-0 mt-0.5">
                <ArrowLeft className="size-4" />
              </Button>
            </Link>
            <div>
              <h1 className="text-2xl font-semibold tracking-tight">New Migration</h1>
              <p className="text-sm text-muted-foreground mt-1">Write a schema migration to version control</p>
            </div>
          </div>

          <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
            {/* Form */}
            <Card className="lg:col-span-2">
              <CardHeader className="border-0">
                <CardTitle className="text-base flex items-center gap-2">
                  <FileCode2 className="size-4" />
                  Migration
                </CardTitle>
                <CardDescription>Describe the change and write the SQL</CardDescription>
              </CardHeader>
              <CardContent className="pt-0 space-y-5">
                <div className="grid gap-2">
                  <Label htmlFor="name">Migration Title</Label>
                  <Input
                    id="name"
                    placeholder="e.g. Add users table composite index"
                    className="h-11"
                    value={form.title}
                    onChange={set("title")}
                  />
                </div>
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                  <div className="grid gap-2">
                    <Label htmlFor="version">Version</Label>
                    <Input
                      id="version"
                      placeholder="v1.3.0"
                      className="h-11 font-mono"
                      value={form.version}
                      onChange={set("version")}
                    />
                  </div>
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="summary">Description</Label>
                  <Textarea
                    id="summary"
                    placeholder="What is changing and why..."
                    className="min-h-[90px]"
                    value={form.description}
                    onChange={set("description")}
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="sql">SQL (up)</Label>
                  <Textarea
                    id="sql"
                    placeholder={"ALTER TABLE public.users\n  ADD COLUMN example_column text NOT NULL DEFAULT '';\n"}
                    className="min-h-[220px] font-mono text-xs leading-relaxed"
                    value={form.upSql}
                    onChange={set("upSql")}
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="down">SQL (down) — optional</Label>
                  <Textarea
                    id="down"
                    placeholder={"ALTER TABLE public.users\n  DROP COLUMN example_column;\n"}
                    className="min-h-[120px] font-mono text-xs leading-relaxed"
                    value={form.downSql}
                    onChange={set("downSql")}
                  />
                </div>
              </CardContent>
            </Card>

            {/* Sidebar */}
            <div className="space-y-6">
              <Card>
                <CardHeader className="border-0">
                  <CardTitle className="text-base">Submission</CardTitle>
                </CardHeader>
                <CardContent className="pt-0">
                  <div className="grid gap-4">
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">Status</span>
                      <span className="text-sm font-medium">Draft</span>
                    </div>
                    <Separator />
                    <div className="rounded-md bg-muted px-3 py-2.5">
                      <p className="text-xs text-muted-foreground">
                        Creating a migration saves it as a draft. You can run it against a connection afterwards.
                      </p>
                    </div>
                    <Button
                      className="h-11 gap-2 w-full"
                      onClick={submit}
                      disabled={!canSubmit || createMigration.isPending}
                    >
                      {createMigration.isPending ? (
                        <>
                          <Loader2 className="size-4 animate-spin" />
                          Creating…
                        </>
                      ) : (
                        <>
                          <Send className="size-4" />
                          Create Migration
                        </>
                      )}
                    </Button>
                  </div>
                  {createMigration.isError && (
                    <p className="mt-4 text-sm text-red-500">{getApiErrorMessage(createMigration.error)}</p>
                  )}
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
