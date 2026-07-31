"use client";

import { useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, CheckCheck, CornerUpLeft, XCircle, Database, Table2, GitBranch, Clock, Search, GitPullRequest } from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbLink,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Separator } from "@/components/ui/separator";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { NotificationsPopover } from "@/components/notifications-popover";
import { reviews, statusConfig, priorityConfig, type ReviewItem } from "@/lib/reviews-data";

export default function ReviewDetailPage() {
  const params = useParams();
  const id = params.id as string;
  const review = reviews.find((r) => r.id === id);
  const [status, setStatus] = useState<ReviewItem["status"] | null>(null);

  if (!review) {
    return (
      <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
        <AppSidebar />
        <SidebarInset>
          <div className="flex flex-col items-center justify-center h-full gap-4 p-8">
            <GitPullRequest className="size-12 text-muted-foreground/40" />
            <h2 className="text-xl font-semibold">Review not found</h2>
            <p className="text-sm text-muted-foreground">The review you are looking for does not exist.</p>
            <Link href="/reviews">
              <Button variant="outline">Back to Reviews</Button>
            </Link>
          </div>
        </SidebarInset>
      </SidebarProvider>
    );
  }

  const currentStatus = status ?? review.status;
  const statusInfo = statusConfig[currentStatus];
  const priority = priorityConfig[review.priority];

  return (
    <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
      <AppSidebar />
      <SidebarInset>
        <header className="sticky top-0 flex h-14 shrink-0 items-center gap-2 border-b bg-background px-4">
          <SidebarTrigger className="-ml-1 size-9" />
          <Separator orientation="vertical" className="mr-2 data-[orientation=vertical]:h-5" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem><BreadcrumbLink href="/reviews">Reviews</BreadcrumbLink></BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem><BreadcrumbPage className="max-w-[300px] truncate">{review.title}</BreadcrumbPage></BreadcrumbItem>
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
          {/* Review header */}
          <div className="flex flex-col gap-4">
            <div className="flex items-start gap-4">
              <Link href="/reviews">
                <Button variant="ghost" size="icon" className="size-10 shrink-0 mt-0.5">
                  <ArrowLeft className="size-4" />
                </Button>
              </Link>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2.5 flex-wrap">
                  <div className={`size-2.5 rounded-full ${statusInfo.dot}`} />
                  <h1 className="text-2xl font-semibold tracking-tight">{review.title}</h1>
                  <Badge variant={statusInfo.badge} className="text-[11px]">{statusInfo.label}</Badge>
                </div>
                <div className="flex items-center gap-2 mt-1.5 text-sm text-muted-foreground flex-wrap">
                  <Avatar className="size-5">
                    <AvatarFallback className="text-[8px]">{review.initials}</AvatarFallback>
                  </Avatar>
                  <span className="text-foreground font-medium">{review.author}</span>
                  <span>requested this review</span>
                  <span>·</span>
                  <span>{review.time}</span>
                </div>
              </div>
              {(currentStatus === "pending" || currentStatus === "changes") && (
                <div className="flex items-center gap-2 shrink-0">
                  <Button variant="outline" className="h-10 gap-2" onClick={() => setStatus("changes")}>
                    <CornerUpLeft className="size-4" />
                    Request Changes
                  </Button>
                  <Button variant="outline" className="h-10 gap-2 text-destructive" onClick={() => setStatus("rejected")}>
                    <XCircle className="size-4" />
                    Reject
                  </Button>
                  <Button className="h-10 gap-2" onClick={() => setStatus("approved")}>
                    <CheckCheck className="size-4" />
                    Approve
                  </Button>
                </div>
              )}
            </div>
          </div>

          {/* Main grid */}
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
            {/* Left column */}
            <div className="space-y-6 lg:col-span-2">
              {/* Schema changes */}
              <Card>
                <CardHeader className="border-0">
                  <CardTitle className="text-base flex items-center gap-2">
                    <GitBranch className="size-4" />
                    Schema Changes
                  </CardTitle>
                </CardHeader>
                <CardContent className="pt-0">
                  <div className="flex flex-col gap-2">
                    {review.changes.map((change, i) => (
                      <div key={i} className="flex items-start gap-3 rounded-md border px-3 py-2.5">
                        <span
                          className={`font-mono text-sm leading-5 ${
                            change.action === "add" ? "text-green-600" :
                            change.action === "remove" ? "text-red-600" : "text-amber-600"
                          }`}
                        >
                          {change.action === "add" ? "+" : change.action === "remove" ? "-" : "~"}
                        </span>
                        <div className="min-w-0">
                          <p className="font-mono text-sm">
                            {change.column}
                            <span className="text-muted-foreground"> {change.type}</span>
                          </p>
                          {change.note && <p className="text-xs text-muted-foreground mt-0.5">{change.note}</p>}
                        </div>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>

              {/* Description */}
              <Card>
                <CardHeader className="border-0">
                  <CardTitle className="text-base">Description</CardTitle>
                </CardHeader>
                <CardContent className="pt-0">
                  <p className="text-sm text-muted-foreground leading-relaxed">{review.description}</p>
                </CardContent>
              </Card>

              {/* Timeline */}
              <Card>
                <CardHeader className="border-0">
                  <CardTitle className="text-base flex items-center gap-2">
                    <Clock className="size-4" />
                    Activity
                  </CardTitle>
                </CardHeader>
                <CardContent className="pt-0">
                  <div className="flex items-start gap-3 py-3 border-b last:border-b-0">
                    <div className="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-full bg-muted">
                      <GitPullRequest className="size-4 text-muted-foreground" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm">
                        <span className="font-medium">{review.author}</span> requested a review
                      </p>
                      <p className="text-xs text-muted-foreground mt-0.5">{review.time}</p>
                    </div>
                  </div>
                  {review.reviewers.map((reviewer, i) => (
                    <div key={i} className="flex items-start gap-3 py-3 border-b last:border-b-0">
                      <div
                        className={`mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-full text-white ${
                          reviewer.decision === "approved" ? "bg-green-500" :
                          reviewer.decision === "changes" ? "bg-amber-500" : "bg-muted text-muted-foreground"
                        }`}
                      >
                        {reviewer.decision === "approved" ? <CheckCheck className="size-4" /> :
                         reviewer.decision === "changes" ? <CornerUpLeft className="size-4" /> :
                         <Clock className="size-4" />}
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm">
                          <span className="font-medium">{reviewer.name}</span>{" "}
                          {reviewer.decision === "approved" ? "approved these changes" :
                           reviewer.decision === "changes" ? "requested changes" : "is reviewing"}
                        </p>
                        <p className="text-xs text-muted-foreground mt-0.5">
                          {reviewer.decision === "approved" ? "2 hours ago" :
                           reviewer.decision === "changes" ? "1 day ago" : "Not yet"}
                        </p>
                      </div>
                    </div>
                  ))}
                </CardContent>
              </Card>
            </div>

            {/* Right column */}
            <div className="space-y-6">
              {/* Details */}
              <Card>
                <CardHeader className="border-0">
                  <CardTitle className="text-base">Details</CardTitle>
                </CardHeader>
                <CardContent className="pt-0">
                  <div className="space-y-3">
                    <div className="flex items-center gap-3 text-sm">
                      <Database className="size-4 text-muted-foreground shrink-0" />
                      <div className="min-w-0">
                        <p className="text-xs text-muted-foreground">Database</p>
                        <p className="font-mono text-xs truncate">{review.database}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-3 text-sm">
                      <Table2 className="size-4 text-muted-foreground shrink-0" />
                      <div className="min-w-0">
                        <p className="text-xs text-muted-foreground">Table</p>
                        <p className="font-mono text-xs">{review.table}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-3 text-sm">
                      <GitBranch className="size-4 text-muted-foreground shrink-0" />
                      <div className="min-w-0">
                        <p className="text-xs text-muted-foreground">Migration</p>
                        <p className="font-mono text-xs">{review.version}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-3 text-sm">
                      <GitPullRequest className="size-4 text-muted-foreground shrink-0" />
                      <div className="min-w-0">
                        <p className="text-xs text-muted-foreground">Project</p>
                        <Link href="/projects" className="text-sm font-medium hover:underline">
                          {review.project}
                        </Link>
                      </div>
                    </div>
                    <Separator />
                    <div className="flex items-center justify-between">
                      <span className="text-xs text-muted-foreground">Priority</span>
                      <Badge variant={priority.variant} className="text-[10px] px-1.5 py-0">{priority.label}</Badge>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* Reviewers */}
              <Card>
                <CardHeader className="border-0">
                  <CardTitle className="text-base">Reviewers</CardTitle>
                </CardHeader>
                <CardContent className="pt-0">
                  <div className="space-y-3">
                    {review.reviewers.map((reviewer, i) => (
                      <div key={i} className="flex items-center gap-3">
                        <Avatar className="size-8">
                          <AvatarFallback className="text-xs">{reviewer.initials}</AvatarFallback>
                        </Avatar>
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium truncate">{reviewer.name}</p>
                          <p className="text-xs text-muted-foreground">
                            {reviewer.decision === "approved" ? "Approved" :
                             reviewer.decision === "changes" ? "Requested changes" : "Waiting for review"}
                          </p>
                        </div>
                        {reviewer.decision === "approved" && <CheckCheck className="size-4 text-green-500 shrink-0" />}
                        {reviewer.decision === "changes" && <CornerUpLeft className="size-4 text-amber-500 shrink-0" />}
                        {reviewer.decision === "pending" && <Clock className="size-4 text-muted-foreground shrink-0" />}
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
