"use client"

import * as React from "react"
import Link from "next/link"
import Image from "next/image"
import { usePathname } from "next/navigation"
import { useQueries } from "@tanstack/react-query"
import { Database, FolderKanban, LayoutDashboard, Settings, Users, CheckCircle2, AlertCircle, Bell } from "lucide-react"

import { NavUser } from "@/components/nav-user"
import { useAuth } from "@/lib/api/auth-provider"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { useProjects, useMembers } from "@/lib/api/hooks/use-projects"
import { useSchemas } from "@/lib/api/hooks/use-schemas"
import { useAuditEntries } from "@/lib/api/hooks/use-audit"

const navMain = [
  { title: "Dashboard", url: "/dashboard", icon: LayoutDashboard },
  { title: "Projects", url: "/projects", icon: FolderKanban },
  { title: "Schemas", url: "/schemas", icon: Database },
  { title: "Team", url: "/team", icon: Users },
  { title: "Settings", url: "/settings", icon: Settings },
]

function relativeTime(iso: string | undefined): string {
  if (!iso) return "just now"
  const then = new Date(iso).getTime()
  if (Number.isNaN(then)) return "just now"
  const mins = Math.floor((Date.now() - then) / 60_000)
  if (mins < 1) return "just now"
  if (mins < 60) return `${mins}m ago`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}h ago`
  return `${Math.floor(hours / 24)}d ago`
}

function initialsOf(name: string): string {
  return name
    .split(/[\s@._-]+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? "")
    .join("") || "?"
}

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const pathname = usePathname()
  const { user } = useAuth()
  const [activeItem, setActiveItem] = React.useState(
    navMain.find((item) => pathname.startsWith(item.url)) || navMain[0]
  )
  useSidebar()

  const { data: projects = [] } = useProjects()
  const firstProjectId = projects[0]?.id

  const { data: schemas = [] } = useSchemas(firstProjectId)
  const { data: members = [] } = useMembers(firstProjectId)

  const projectQueries = useQueries({
    queries: projects.map((p) => ({
      queryKey: ["projects", p.id, "members"],
      queryFn: async () => {
        const res = await (await import("@/lib/api/clients")).projectClient.listMembers({
          projectId: p.id,
        })
        return res.members.length
      },
      enabled: Boolean(p.id),
    })),
  })
  const connectionQueries = useQueries({
    queries: projects.map((p) => ({
      queryKey: ["projects", p.id, "connections"],
      queryFn: async () => {
        const res = await (await import("@/lib/api/clients")).projectClient.listConnections({
          projectId: p.id,
        })
        return res.connections.length
      },
      enabled: Boolean(p.id),
    })),
  })

  const totalMembers = projectQueries.reduce((sum, q) => sum + (q.data ?? 0), 0)
  const totalConnections = connectionQueries.reduce((sum, q) => sum + (q.data ?? 0), 0)
  const { data: auditEntries = [] } = useAuditEntries({})

  const activityItems = auditEntries.slice(0, 4).map((entry) => ({
    action: entry.action || entry.eventType || "Event",
    project: entry.resourceType ? `${entry.resourceType} ${entry.resourceId || ""}`.trim() : "",
    time: relativeTime(entry.createdAt),
    icon: entry.eventType?.toLowerCase().includes("drift") ? AlertCircle : CheckCircle2,
    color: entry.eventType?.toLowerCase().includes("drift") ? "text-amber-500" : "text-green-500",
  }))

  return (
    <Sidebar
      collapsible="icon"
      className="overflow-hidden *:data-[sidebar=sidebar]:flex-row"
      {...props}
    >
      <Sidebar
        collapsible="none"
        className="w-[calc(var(--sidebar-width-icon)+1px)]! border-r"
      >
        <SidebarHeader>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton size="lg" asChild className="md:h-9 md:p-0">
                <Link href="/dashboard">
                  <Image
                    src="/logo.svg"
                    alt="Mainline"
                    width={94}
                    height={18}
                    className="dark:invert"
                  />
                </Link>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarHeader>
        <SidebarContent>
          <SidebarGroup>
            <SidebarGroupContent className="px-1.5 md:px-0">
              <SidebarMenu>
                {navMain.map((item) => (
                  <SidebarMenuItem key={item.title}>
                    <SidebarMenuButton
                      tooltip={{ children: item.title, hidden: false }}
                      onClick={() => setActiveItem(item)}
                      isActive={pathname.startsWith(item.url)}
                      className="px-2.5 md:px-2"
                      asChild
                    >
                      <Link href={item.url} prefetch>
                        <item.icon />
                        <span>{item.title}</span>
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>
        <SidebarFooter>
          <NavUser
            user={{
              name: user?.displayName || "Guest",
              email: user?.email || "",
              avatar: user?.avatarUrl || "",
            }}
          />
        </SidebarFooter>
      </Sidebar>

      <Sidebar collapsible="none" className="hidden flex-1 md:flex">
        <SidebarHeader className="gap-3.5 border-b p-4">
          <div className="flex w-full items-center justify-between">
            <div className="text-base font-medium text-foreground">
              {activeItem?.title === "Dashboard" ? "Overview" : activeItem?.title || "Overview"}
            </div>
          </div>
        </SidebarHeader>
        <SidebarContent>
          {/* Dashboard sidebar content */}
          {activeItem?.title === "Dashboard" && (
            <SidebarGroup className="px-0">
              <SidebarGroupContent>
                <div className="px-4 py-3 text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Recent Activity
                </div>
                {activityItems.length === 0 && (
                  <p className="px-4 py-3 text-sm text-muted-foreground">No activity yet.</p>
                )}
                {activityItems.map((item) => (
                  <div key={item.action + item.time} className="flex items-start gap-3 border-b p-4 text-sm last:border-b-0 hover:bg-sidebar-accent">
                    <item.icon className={`size-4 mt-0.5 shrink-0 ${item.color}`} />
                    <div className="min-w-0">
                      <div className="font-medium truncate">{item.action}</div>
                      <div className="flex items-center gap-2 text-xs text-muted-foreground mt-0.5">
                        {item.project && <span>{item.project}</span>}
                        <span>{item.time}</span>
                      </div>
                    </div>
                  </div>
                ))}
                <div className="px-4 py-3 text-xs font-medium text-muted-foreground uppercase tracking-wider border-t">
                  Quick Stats
                </div>
                <div className="grid grid-cols-2 gap-3 p-4">
                  <div className="rounded-lg bg-sidebar-accent/50 p-3">
                    <p className="text-lg font-semibold">{projects.length}</p>
                    <p className="text-xs text-muted-foreground">Projects</p>
                  </div>
                  <div className="rounded-lg bg-sidebar-accent/50 p-3">
                    <p className="text-lg font-semibold">{totalConnections}</p>
                    <p className="text-xs text-muted-foreground">Connections</p>
                  </div>
                  <div className="rounded-lg bg-sidebar-accent/50 p-3">
                    <p className="text-lg font-semibold">{totalMembers}</p>
                    <p className="text-xs text-muted-foreground">Members</p>
                  </div>
                  <div className="rounded-lg bg-sidebar-accent/50 p-3">
                    <p className="text-lg font-semibold">{schemas.length}</p>
                    <p className="text-xs text-muted-foreground">Schemas</p>
                  </div>
                </div>
              </SidebarGroupContent>
            </SidebarGroup>
          )}

          {/* Projects sidebar content */}
          {activeItem?.title === "Projects" && (
            <SidebarGroup className="px-0">
              <SidebarGroupContent>
                <div className="px-4 py-3 text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Project Stats
                </div>
                <div className="grid grid-cols-2 gap-3 p-4">
                  <div className="rounded-lg bg-sidebar-accent/50 p-3">
                    <p className="text-lg font-semibold">{projects.length}</p>
                    <p className="text-xs text-muted-foreground">Total</p>
                  </div>
                  <div className="rounded-lg bg-sidebar-accent/50 p-3">
                    <p className="text-lg font-semibold">{totalMembers}</p>
                    <p className="text-xs text-muted-foreground">Members</p>
                  </div>
                  <div className="rounded-lg bg-sidebar-accent/50 p-3">
                    <p className="text-lg font-semibold">{totalConnections}</p>
                    <p className="text-xs text-muted-foreground">Connections</p>
                  </div>
                  <div className="rounded-lg bg-sidebar-accent/50 p-3">
                    <p className="text-lg font-semibold">{schemas.length}</p>
                    <p className="text-xs text-muted-foreground">Schemas</p>
                  </div>
                </div>
                <div className="px-4 py-3 text-xs font-medium text-muted-foreground uppercase tracking-wider border-t">
                  Recent Projects
                </div>
                {projects.length === 0 && (
                  <p className="px-4 py-3 text-sm text-muted-foreground">No projects yet. Create one!</p>
                )}
                {projects.slice(0, 5).map((project) => (
                  <Link
                    key={project.id}
                    href={`/projects/${project.id}`}
                    prefetch
                    className="flex flex-col gap-1 border-b p-4 text-sm leading-tight last:border-b-0 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
                  >
                    <div className="font-medium">{project.name}</div>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                      <span className="inline-block size-1.5 rounded-full bg-green-500" />
                      <span>{relativeTime(project.createdAt)}</span>
                    </div>
                  </Link>
                ))}
              </SidebarGroupContent>
            </SidebarGroup>
          )}

          {/* Schemas sidebar content */}
          {activeItem?.title === "Schemas" && (
            <SidebarGroup className="px-0">
              <SidebarGroupContent>
                <div className="px-4 py-3 text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  All Schemas
                </div>
                {schemas.length === 0 && (
                  <p className="px-4 py-3 text-sm text-muted-foreground">
                    {firstProjectId ? "No schemas yet." : "No projects yet."}
                  </p>
                )}
                {schemas.map((schema) => (
                  <div key={schema.schemaName} className="flex items-center justify-between border-b p-4 text-sm last:border-b-0 hover:bg-sidebar-accent">
                    <div className="flex items-center gap-3">
                      <Database className="size-4 text-muted-foreground" />
                      <div>
                      <div className="font-medium">{schema.schemaName}</div>
                      <div className="text-xs text-muted-foreground">
                        {schema.currentVersionId ? `version ${schema.currentVersionId.slice(0, 8)}` : "no version"}
                      </div>
                      </div>
                    </div>
                    <span className="text-xs text-muted-foreground">{relativeTime(schema.updatedAt)}</span>
                  </div>
                ))}
              </SidebarGroupContent>
            </SidebarGroup>
          )}

          {/* Team sidebar content */}
          {activeItem?.title === "Team" && (
            <SidebarGroup className="px-0">
              <SidebarGroupContent>
                <div className="px-4 py-3 text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Team Members
                </div>
                {members.length === 0 && (
                  <p className="px-4 py-3 text-sm text-muted-foreground">No team members yet.</p>
                )}
                {members.map((member) => (
                  <div key={member.userId} className="flex items-center gap-3 border-b p-4 text-sm last:border-b-0 hover:bg-sidebar-accent">
                    <Avatar className="size-8">
                      <AvatarFallback className="text-xs">{initialsOf(member.displayName || member.email)}</AvatarFallback>
                    </Avatar>
                    <div className="min-w-0">
                      <div className="font-medium truncate">{member.displayName || member.email}</div>
                      <div className="text-xs text-muted-foreground capitalize">{member.role}</div>
                    </div>
                  </div>
                ))}
              </SidebarGroupContent>
            </SidebarGroup>
          )}

          {/* Settings sidebar content */}
          {activeItem?.title === "Settings" && (
            <SidebarGroup className="px-0">
              <SidebarGroupContent>
                <Link href="/settings" className="flex items-center gap-3 border-b p-4 text-sm hover:bg-sidebar-accent" prefetch>
                  <Settings className="size-4 text-muted-foreground" />
                  <span>Profile Settings</span>
                </Link>
                <Link href="/settings/connections" className="flex items-center gap-3 border-b p-4 text-sm hover:bg-sidebar-accent" prefetch>
                  <Database className="size-4 text-muted-foreground" />
                  <span>Database Connections</span>
                </Link>
                <Link href="/team" className="flex items-center gap-3 border-b p-4 text-sm hover:bg-sidebar-accent" prefetch>
                  <Users className="size-4 text-muted-foreground" />
                  <span>Team Management</span>
                </Link>
                <Link href="/settings" className="flex items-center gap-3 border-b p-4 text-sm hover:bg-sidebar-accent" prefetch>
                  <Bell className="size-4 text-muted-foreground" />
                  <span>Notifications</span>
                </Link>
              </SidebarGroupContent>
            </SidebarGroup>
          )}
        </SidebarContent>
      </Sidebar>
    </Sidebar>
  )
}
