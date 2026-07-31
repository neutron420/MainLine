"use client";

import { useState, useMemo, useEffect } from "react";
import { Plus, Search, ChevronLeft, ChevronRight, GitPullRequest } from "lucide-react";
import Link from "next/link";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
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
  DropdownMenuCheckboxItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Separator } from "@/components/ui/separator";
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
import { reviews, statusConfig, priorityConfig, projectOptions } from "@/lib/reviews-data";

const PER_PAGE = 8;

const tabs: { key: string; label: string }[] = [
  { key: "all", label: "All" },
  { key: "pending", label: "Pending" },
  { key: "approved", label: "Approved" },
  { key: "changes", label: "Changes" },
  { key: "rejected", label: "Rejected" },
];

export default function ReviewsPage() {
  const [search, setSearch] = useState("");
  const [statusTab, setStatusTab] = useState("all");
  const [projectFilter, setProjectFilter] = useState("all");
  const [page, setPage] = useState(1);
  const [createOpen, setCreateOpen] = useState(false);

  const statusCounts = useMemo(() => {
    const counts: Record<string, number> = { all: reviews.length };
    for (const r of reviews) counts[r.status] = (counts[r.status] || 0) + 1;
    return counts;
  }, []);

  const filtered = useMemo(() => {
    return reviews.filter((r) => {
      const matchSearch =
        r.title.toLowerCase().includes(search.toLowerCase()) ||
        r.table.toLowerCase().includes(search.toLowerCase());
      const matchStatus = statusTab === "all" || r.status === statusTab;
      const matchProject = projectFilter === "all" || r.project === projectFilter;
      return matchSearch && matchStatus && matchProject;
    });
  }, [search, statusTab, projectFilter]);

  const totalPages = Math.max(1, Math.ceil(filtered.length / PER_PAGE));
  const safePage = Math.min(page, totalPages);
  const paginated = useMemo(() => {
    const start = (safePage - 1) * PER_PAGE;
    return filtered.slice(start, start + PER_PAGE);
  }, [filtered, safePage]);

  useEffect(() => {
    if (page > totalPages) setPage(1);
  }, [page, totalPages]);

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
              <BreadcrumbItem><BreadcrumbPage>Reviews</BreadcrumbPage></BreadcrumbItem>
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
          {/* Header + New Review */}
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-semibold tracking-tight">Reviews</h1>
              <p className="text-sm text-muted-foreground mt-1">Schema change reviews across projects</p>
            </div>
            <Dialog open={createOpen} onOpenChange={setCreateOpen}>
              <DialogTrigger asChild>
                <Tooltip delay={0}>
                  <Button className="h-11 gap-2">
                    <Plus className="size-4" />
                    New Review
                  </Button>
                  <Tooltip.Content>
                    <p>Request a schema review</p>
                  </Tooltip.Content>
                </Tooltip>
              </DialogTrigger>
              <DialogContent className="sm:max-w-[500px]">
                <DialogHeader>
                  <DialogTitle>New Review</DialogTitle>
                  <DialogDescription>
                    Submit a schema change for review by your team.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-5 py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="title">Review Title</Label>
                    <Input id="title" placeholder="e.g. Add users table composite index" className="h-11" />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="project">Project</Label>
                    <Select>
                      <SelectTrigger id="project" className="h-11">
                        <SelectValue placeholder="Select project" />
                      </SelectTrigger>
                      <SelectContent>
                        {projectOptions.map((p) => (
                          <SelectItem key={p} value={p.toLowerCase().replace(/ /g, "-")}>{p}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="grid gap-2">
                      <Label htmlFor="table">Table</Label>
                      <Input id="table" placeholder="users" className="h-11" />
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="version">Migration</Label>
                      <Input id="version" placeholder="v1.2.0" className="h-11" />
                    </div>
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="desc">Description</Label>
                    <Textarea id="desc" placeholder="What is changing and why..." className="min-h-[80px]" />
                  </div>
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
                  <Button onClick={() => setCreateOpen(false)}>Submit Review</Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </div>

          {/* Status tabs */}
          <Tabs value={statusTab} onValueChange={(v) => { setStatusTab(v); setPage(1); }}>
            <TabsList variant="line" className="w-full justify-start gap-4">
              {tabs.map((tab) => (
                <TabsTrigger key={tab.key} value={tab.key} className="px-1 gap-1.5">
                  {tab.label}
                  <span className="text-xs text-muted-foreground">{statusCounts[tab.key] || 0}</span>
                </TabsTrigger>
              ))}
            </TabsList>
          </Tabs>

          {/* Search + Filter */}
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
            <div className="relative flex-1 max-w-md">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
              <Input
                placeholder="Search reviews..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-9 h-11"
              />
            </div>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" className="h-11">
                  {projectFilter === "all" ? "Project" : projectFilter}
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                <DropdownMenuLabel>Filter by Project</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuCheckboxItem checked={projectFilter === "all"} onCheckedChange={() => setProjectFilter("all")}>All Projects</DropdownMenuCheckboxItem>
                {projectOptions.map((p) => (
                  <DropdownMenuCheckboxItem key={p} checked={projectFilter === p} onCheckedChange={() => setProjectFilter(p)}>{p}</DropdownMenuCheckboxItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>

          {/* Results count */}
          <p className="text-sm text-muted-foreground">
            Showing {paginated.length} of {filtered.length} review{filtered.length !== 1 ? "s" : ""}
          </p>

          {/* Review list */}
          <div className="flex flex-col gap-4">
            {paginated.map((review) => {
              const status = statusConfig[review.status];
              const priority = priorityConfig[review.priority];
              return (
                <Link key={review.id} href={`/reviews/${review.id}`}>
                  <Card className="cursor-pointer transition-all hover:shadow-md hover:border-primary/50">
                    <CardContent className="p-5">
                      <div className="flex items-start justify-between gap-4">
                        <div className="flex items-center gap-2.5 min-w-0">
                          <div className={`size-2 shrink-0 rounded-full ${status.dot}`} />
                          <h3 className="font-medium truncate">{review.title}</h3>
                          <Badge variant={status.badge} className="text-[10px] px-1.5 py-0 shrink-0">{status.label}</Badge>
                        </div>
                        <span className="text-xs text-muted-foreground shrink-0">{review.time}</span>
                      </div>
                      <p className="text-xs text-muted-foreground mt-1.5 pl-4.5">
                        {review.project} · {review.table} · {review.version}
                      </p>
                      <div className="flex items-center gap-3 mt-4 pt-3.5 border-t">
                        <Avatar className="size-6">
                          <AvatarFallback className="text-[9px]">{review.initials}</AvatarFallback>
                        </Avatar>
                        <span className="text-xs text-muted-foreground">{review.author}</span>
                        <div className="ml-auto flex items-center gap-2">
                          <Badge variant={priority.variant} className="text-[10px] px-1.5 py-0">{priority.label}</Badge>
                          {(review.status === "pending" || review.status === "changes") && (
                            <Button size="sm" className="h-8">Review</Button>
                          )}
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </Link>
              );
            })}
          </div>

          {/* Pagination */}
          {totalPages > 1 && filtered.length > 0 && (
            <div className="flex items-center justify-center gap-2 pt-2">
              <Tooltip delay={0}>
                <Button
                  variant="outline"
                  size="icon"
                  className="size-10"
                  disabled={page === 1}
                  onClick={() => setPage((p) => p - 1)}
                >
                  <ChevronLeft className="size-4" />
                </Button>
                <Tooltip.Content>
                  <p>Previous page</p>
                </Tooltip.Content>
              </Tooltip>
              {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
                <Tooltip key={p} delay={0}>
                  <Button
                    variant={p === page ? "default" : "outline"}
                    size="icon"
                    className="size-10"
                    onClick={() => setPage(p)}
                  >
                    {p}
                  </Button>
                  <Tooltip.Content>
                    <p>Page {p}</p>
                  </Tooltip.Content>
                </Tooltip>
              ))}
              <Tooltip delay={0}>
                <Button
                  variant="outline"
                  size="icon"
                  className="size-10"
                  disabled={page === totalPages}
                  onClick={() => setPage((p) => p + 1)}
                >
                  <ChevronRight className="size-4" />
                </Button>
                <Tooltip.Content>
                  <p>Next page</p>
                </Tooltip.Content>
              </Tooltip>
            </div>
          )}

          {/* Empty state */}
          {filtered.length === 0 && (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <GitPullRequest className="size-12 text-muted-foreground/40 mb-4" />
              <h3 className="text-lg font-medium">No reviews found</h3>
              <p className="text-sm text-muted-foreground mt-1">Try adjusting your search or filters</p>
            </div>
          )}
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
