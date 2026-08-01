"use client";

import { useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Search, FileCode2, Save, Send } from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
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
import { environments } from "@/lib/migrations-data";

export default function CreateMigrationPage() {
  const params = useParams();
  const projectId = params.id as string;
  const [submitted, setSubmitted] = useState(false);

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
              <BreadcrumbItem><BreadcrumbLink href={`/projects/${projectId}`}>Project</BreadcrumbLink></BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem><BreadcrumbPage>New Migration</BreadcrumbPage></BreadcrumbItem>
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
          {/* Header */}
          <div className="flex items-start gap-4">
            <Link href={`/projects/${projectId}`}>
              <Button variant="ghost" size="icon" className="size-10 shrink-0 mt-0.5">
                <ArrowLeft className="size-4" />
              </Button>
            </Link>
            <div>
              <h1 className="text-2xl font-semibold tracking-tight">New Migration</h1>
              <p className="text-sm text-muted-foreground mt-1">Write and submit a schema migration for review</p>
            </div>
          </div>

          {submitted ? (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-16 text-center">
                <Send className="size-12 text-green-500 mb-4" />
                <h2 className="text-lg font-semibold">Migration submitted for review</h2>
                <p className="text-sm text-muted-foreground mt-1">
                  Your team has been notified. You can track the review on the project page.
                </p>
                <div className="flex items-center gap-2 mt-6">
                  <Link href={`/projects/${projectId}/migrations/m2`}>
                    <Button variant="outline">View Example Migration</Button>
                  </Link>
                  <Link href={`/projects/${projectId}`}>
                    <Button>Back to Project</Button>
                  </Link>
                </div>
              </CardContent>
            </Card>
          ) : (
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
                    <Label htmlFor="name">Migration Name</Label>
                    <Input id="name" placeholder="e.g. Add users table composite index" className="h-11" />
                  </div>
                  <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
                    <div className="grid gap-2">
                      <Label htmlFor="version">Version</Label>
                      <Input id="version" defaultValue="v1.3.0" className="h-11 font-mono" />
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="table">Table</Label>
                      <Select>
                        <SelectTrigger id="table" className="h-11">
                          <SelectValue placeholder="Select table" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="users">users</SelectItem>
                          <SelectItem value="teams">teams</SelectItem>
                          <SelectItem value="memberships">memberships</SelectItem>
                          <SelectItem value="sessions">sessions</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="env">Environment</Label>
                      <Select defaultValue="Staging">
                        <SelectTrigger id="env" className="h-11">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {environments.map((env) => (
                            <SelectItem key={env} value={env}>{env}</SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="summary">Summary</Label>
                    <Textarea id="summary" placeholder="What is changing and why..." className="min-h-[90px]" />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="sql">SQL</Label>
                    <Textarea
                      id="sql"
                      placeholder={"ALTER TABLE public.users\n  ADD COLUMN example_column text NOT NULL DEFAULT '';\n"}
                      className="min-h-[220px] font-mono text-xs leading-relaxed"
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
                        <span className="text-muted-foreground">Author</span>
                        <span className="font-medium">R.K Singh</span>
                      </div>
                      <div className="flex items-center justify-between text-sm">
                        <span className="text-muted-foreground">Status</span>
                        <span className="text-sm font-medium">Draft</span>
                      </div>
                      <Separator />
                      <div className="rounded-md bg-muted px-3 py-2.5">
                        <p className="text-xs text-muted-foreground">
                          Submitting creates a migration record and notifies reviewers on your team.
                        </p>
                      </div>
                      <div className="grid gap-2">
                        <Button className="h-11 gap-2 w-full" onClick={() => setSubmitted(true)}>
                          <Send className="size-4" />
                          Submit for Review
                        </Button>
                        <Button variant="outline" className="h-11 gap-2 w-full">
                          <Save className="size-4" />
                          Save as Draft
                        </Button>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </div>
            </div>
          )}
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
