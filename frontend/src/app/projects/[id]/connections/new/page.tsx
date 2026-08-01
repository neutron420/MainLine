"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import { useState } from "react";
import { ArrowLeft, Search, Loader2, CheckCircle2, XCircle, PlugZap, Database } from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
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
import { Tooltip } from "@heroui/react";

type TestState = "idle" | "testing" | "success" | "error";

export default function CreateConnectionPage() {
  const params = useParams();
  const projectId = params.id as string;

  const [testState, setTestState] = useState<TestState>("idle");
  const [created, setCreated] = useState(false);
  const [env, setEnv] = useState("Development");

  const runTest = () => {
    if (testState === "testing") return;
    setTestState("testing");
    setTimeout(() => setTestState("success"), 1800);
  };

  if (created) {
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
              SchemaHub is now monitoring <span className="font-mono text-foreground">mainline_dev</span>.
              Initial schema sync will complete shortly.
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
                  <Input placeholder="e.g. Production" defaultValue="Development" />
                </div>
                <div className="space-y-1.5">
                  <Label>Environment</Label>
                  <Select value={env} onValueChange={setEnv}>
                    <SelectTrigger className="h-9">
                      <SelectValue placeholder="Select environment" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="Production">Production</SelectItem>
                      <SelectItem value="Staging">Staging</SelectItem>
                      <SelectItem value="Development">Development</SelectItem>
                      <SelectItem value="QA">QA</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-1.5">
                  <Label>Host</Label>
                  <Input placeholder="ep-xxxx.aws.neon.tech" defaultValue="localhost" className="font-mono" />
                </div>
                <div className="space-y-1.5">
                  <Label>Port</Label>
                  <Input placeholder="5432" defaultValue="5432" className="font-mono" />
                </div>
                <div className="space-y-1.5">
                  <Label>Database</Label>
                  <Input placeholder="mydb" defaultValue="mainline_dev" className="font-mono" />
                </div>
                <div className="space-y-1.5">
                  <Label>Username</Label>
                  <Input placeholder="postgres" defaultValue="postgres" className="font-mono" />
                </div>
                <div className="space-y-1.5 sm:col-span-2">
                  <Label>Password</Label>
                  <Input type="password" placeholder="••••••••••••" className="font-mono" />
                </div>
                <div className="space-y-1.5 sm:col-span-2">
                  <Label>SSL mode</Label>
                  <Select defaultValue="prefer">
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
                <Button
                  variant={testState === "error" ? "destructive" : "outline"}
                  onClick={runTest}
                  disabled={testState === "testing"}
                  className="h-9 gap-2"
                >
                  {testState === "testing" ? (
                    <>
                      <Loader2 className="size-4 animate-spin" />
                      Testing…
                    </>
                  ) : (
                    <>
                      <PlugZap className="size-4" />
                      Test Connection
                    </>
                  )}
                </Button>
                {testState === "success" && (
                  <span className="flex items-center gap-1.5 text-sm text-emerald-500">
                    <CheckCircle2 className="size-4" />
                    Connected successfully
                  </span>
                )}
                {testState === "error" && (
                  <span className="flex items-center gap-1.5 text-sm text-red-500">
                    <XCircle className="size-4" />
                    Connection failed
                  </span>
                )}
                <div className="ml-auto flex items-center gap-3">
                  <Link href={`/projects/${projectId}/connections`}>
                    <Button variant="ghost" className="h-9">Cancel</Button>
                  </Link>
                  <Button
                    className="h-9 gap-2"
                    onClick={() => setCreated(true)}
                    disabled={testState !== "success"}
                  >
                    <PlugZap className="size-4" />
                    Add Connection
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>

          {env === "Production" && (
            <div className="flex items-start gap-2.5 rounded-lg border border-amber-500/40 bg-amber-500/5 p-3.5 text-sm">
              <Badge variant="secondary" className="bg-amber-500/15 text-amber-600 text-[10px] px-1.5 py-0 shrink-0 mt-0.5">
                Warning
              </Badge>
              <p className="text-muted-foreground">
                You are connecting to a <span className="font-medium text-foreground">Production</span> environment.
                Schema changes here will require approval from at least one other admin.
              </p>
            </div>
          )}
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
