"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import { useState } from "react";
import { ArrowLeft, Search, Loader2, CheckCircle2, XCircle, PlugZap, Database } from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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
import { useCreateConnection } from "@/lib/api/hooks/use-connections";
import { getApiErrorMessage } from "@/lib/api/errors";

export default function CreateConnectionPage() {
  const params = useParams();
  const projectId = params.id as string;

  const createConnection = useCreateConnection();
  const createdName = createConnection.data?.name;

  const [form, setForm] = useState({
    name: "",
    host: "",
    port: "5432",
    databaseName: "",
    username: "",
    password: "",
    sslMode: "prefer",
  });

  const set = (key: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setForm((prev) => ({ ...prev, [key]: e.target.value }));

  const canSubmit =
    form.name.trim() && form.host.trim() && form.databaseName.trim() && form.username.trim();

  const submit = () => {
    if (!canSubmit || createConnection.isPending) return;
    createConnection.mutate({
      projectId,
      name: form.name.trim(),
      host: form.host.trim(),
      port: Number(form.port) || 5432,
      databaseName: form.databaseName.trim(),
      username: form.username.trim(),
      password: form.password,
      sslMode: form.sslMode,
    });
  };

  if (createConnection.isSuccess && createdName) {
    return (
      <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
        <AppSidebar />
        <SidebarInset>
          <div className="flex flex-1 flex-col items-center justify-center gap-4 p-8">
            <div className="flex size-14 items-center justify-center rounded-full bg-emerald-500/10">
              <CheckCircle2 className="size-7 text-emerald-500" />
            </div>
            <h2 className="text-xl font-semibold">Connection added</h2>
            <p className="text-sm text-muted-foreground max-w-sm text-center">
              SchemaHub is now monitoring <span className="font-mono text-foreground">{createdName}</span>.
              Connect it to a schema to start tracking changes.
            </p>
            <div className="flex items-center gap-3">
              <Link href={`/projects/${projectId}/connections`}>
                <Button variant="outline">Back to Connections</Button>
              </Link>
              <Link href={`/projects/${projectId}`}>
                <Button>Go to Project</Button>
              </Link>
            </div>
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
          <SidebarTrigger className="-ml-1 size-9" />
          <Separator orientation="vertical" className="mr-2 data-[orientation=vertical]:h-5" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem><BreadcrumbLink href="/projects">Projects</BreadcrumbLink></BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem><BreadcrumbLink href={`/projects/${projectId}`}>SchemaHub</BreadcrumbLink></BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem><BreadcrumbLink href={`/projects/${projectId}/connections`}>Connections</BreadcrumbLink></BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem><BreadcrumbPage>New</BreadcrumbPage></BreadcrumbItem>
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
            <Link href={`/projects/${projectId}/connections`}>
              <Button variant="ghost" size="icon" className="size-10 shrink-0 mt-0.5">
                <ArrowLeft className="size-4" />
              </Button>
            </Link>
            <div>
              <div className="flex items-center gap-3 flex-wrap">
                <h1 className="text-2xl font-semibold tracking-tight flex items-center gap-2">
                  <PlugZap className="size-6" />
                  New Connection
                </h1>
              </div>
              <p className="text-sm text-muted-foreground mt-1">Link a PostgreSQL database to monitor its schema</p>
            </div>
          </div>

          <Card>
            <CardHeader className="border-0">
              <CardTitle className="text-base flex items-center gap-2">
                <Database className="size-4" />
                Database details
              </CardTitle>
            </CardHeader>
            <CardContent className="pt-0">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <Label>Connection name</Label>
                  <Input placeholder="e.g. Production" value={form.name} onChange={set("name")} />
                </div>
                <div className="space-y-1.5">
                  <Label>Host</Label>
                  <Input
                    placeholder="ep-xxxx.aws.neon.tech"
                    value={form.host}
                    onChange={set("host")}
                    className="font-mono"
                  />
                </div>
                <div className="space-y-1.5">
                  <Label>Port</Label>
                  <Input
                    placeholder="5432"
                    value={form.port}
                    onChange={set("port")}
                    className="font-mono"
                  />
                </div>
                <div className="space-y-1.5">
                  <Label>Database</Label>
                  <Input
                    placeholder="mydb"
                    value={form.databaseName}
                    onChange={set("databaseName")}
                    className="font-mono"
                  />
                </div>
                <div className="space-y-1.5">
                  <Label>Username</Label>
                  <Input
                    placeholder="postgres"
                    value={form.username}
                    onChange={set("username")}
                    className="font-mono"
                  />
                </div>
                <div className="space-y-1.5">
                  <Label>Password</Label>
                  <Input
                    type="password"
                    placeholder="••••••••••••"
                    value={form.password}
                    onChange={set("password")}
                    className="font-mono"
                  />
                </div>
                <div className="space-y-1.5 sm:col-span-2">
                  <Label>SSL mode</Label>
                  <Select
                    value={form.sslMode}
                    onValueChange={(v) => setForm((prev) => ({ ...prev, sslMode: v }))}
                  >
                    <SelectTrigger className="h-9">
                      <SelectValue placeholder="Select SSL mode" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="required">Required</SelectItem>
                      <SelectItem value="prefer">Prefer</SelectItem>
                      <SelectItem value="disable">Disable</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="mt-6 pt-4 border-t flex items-center gap-3">
                <div className="ml-auto flex items-center gap-3">
                  <Link href={`/projects/${projectId}/connections`}>
                    <Button variant="ghost" className="h-9">Cancel</Button>
                  </Link>
                  <Button
                    className="h-9 gap-2"
                    onClick={submit}
                    disabled={!canSubmit || createConnection.isPending}
                  >
                    {createConnection.isPending ? (
                      <>
                        <Loader2 className="size-4 animate-spin" />
                        Adding…
                      </>
                    ) : (
                      <>
                        <PlugZap className="size-4" />
                        Add Connection
                      </>
                    )}
                  </Button>
                </div>
              </div>

              {createConnection.isError && (
                <div className="mt-4 flex items-center gap-2 text-sm text-red-500">
                  <XCircle className="size-4 shrink-0" />
                  {getApiErrorMessage(createConnection.error)}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
