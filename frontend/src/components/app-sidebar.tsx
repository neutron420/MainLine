"use client"

import * as React from "react"
import { useRouter, usePathname } from "next/navigation"
import { Database, FolderKanban, LayoutDashboard, Settings, Users, CheckCircle2, AlertCircle, Bell } from "lucide-react"

import { NavUser } from "@/components/nav-user"
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

const data = {
  user: {
    name: "R.K Singh",
    email: "rk@mainline.dev",
    avatar: "/avatars/default.jpg",
  },
  navMain: [
    { title: "Dashboard", url: "/dashboard", icon: LayoutDashboard },
    { title: "Projects", url: "/projects", icon: FolderKanban },
    { title: "Schemas", url: "/schemas", icon: Database },
    { title: "Team", url: "/team", icon: Users },
    { title: "Settings", url: "/settings", icon: Settings },
  ],
  projectsList: [
    { name: "User Service Schema", updated: "2 hours ago", status: "active" as const },
    { name: "Payment DB Migration", updated: "5 hours ago", status: "inProgress" as const },
    { name: "Analytics Warehouse", updated: "1 day ago", status: "active" as const },
    { name: "Legacy CRM Schema", updated: "3 days ago", status: "onHold" as const },
  ],
  projectStats: {
    total: 12,
    active: 7,
    migrations: 47,
    drift: 1,
  },
  schemas: [
    { name: "users", tables: 12, lastUpdated: "2h ago" },
    { name: "payments", tables: 8, lastUpdated: "5h ago" },
    { name: "analytics", tables: 24, lastUpdated: "1d ago" },
    { name: "inventory", tables: 15, lastUpdated: "3d ago" },
  ],
  teamMembers: [
    { name: "Alice", initials: "AL", role: "Lead Developer" },
    { name: "Bob", initials: "BO", role: "Developer" },
    { name: "Charlie", initials: "CH", role: "Developer" },
    { name: "Diana", initials: "DI", role: "Data Engineer" },
  ],
  recentActivity: [
    { action: "Migration v1.2 deployed", project: "User Service", time: "2h ago", icon: CheckCircle2, color: "text-green-500" },
    { action: "Schema drift detected", project: "Analytics", time: "5h ago", icon: AlertCircle, color: "text-amber-500" },
    { action: "New project created", project: "Search Index", time: "2d ago", icon: FolderKanban, color: "text-blue-500" },
  ],
}

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const router = useRouter()
  const pathname = usePathname()
  const [activeItem, setActiveItem] = React.useState(
    data.navMain.find((item) => pathname.startsWith(item.url)) || data.navMain[0]
  )
  useSidebar()

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
                <a href="/dashboard">
                  <div className="grid flex-1 text-left text-sm leading-tight">
                    <span className="truncate font-semibold">Mainline</span>
                    <span className="truncate text-[10px] text-muted-foreground">Schema Hub</span>
                  </div>
                </a>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarHeader>
        <SidebarContent>
          <SidebarGroup>
            <SidebarGroupContent className="px-1.5 md:px-0">
              <SidebarMenu>
                {data.navMain.map((item) => (
                  <SidebarMenuItem key={item.title}>
                    <SidebarMenuButton
                      tooltip={{ children: item.title, hidden: false }}
                      onClick={() => {
                        setActiveItem(item)
                        router.push(item.url)
                      }}
                      isActive={pathname.startsWith(item.url)}
                      className="px-2.5 md:px-2"
                    >
                      <item.icon />
                      <span>{item.title}</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>
        <SidebarFooter>
          <NavUser user={data.user} />
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
                {data.recentActivity.map((item) => (
                  <div key={item.action} className="flex items-start gap-3 border-b p-4 text-sm last:border-b-0 hover:bg-sidebar-accent">
                    <item.icon className={`size-4 mt-0.5 shrink-0 ${item.color}`} />
                    <div className="min-w-0">
                      <div className="font-medium truncate">{item.action}</div>
                      <div className="flex items-center gap-2 text-xs text-muted-foreground mt-0.5">
                        <span>{item.project}</span>
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
                    <p className="text-lg font-semibold">12</p>
                    <p className="text-xs text-muted-foreground">Projects</p>
                  </div>
                  <div className="rounded-lg bg-sidebar-accent/50 p-3">
                    <p className="text-lg font-semibold">47</p>
                    <p className="text-xs text-muted-foreground">Migrations</p>
                  </div>
                  <div className="rounded-lg bg-sidebar-accent/50 p-3">
                    <p className="text-lg font-semibold">24</p>
                    <p className="text-xs text-muted-foreground">Members</p>
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
                    <p className="text-lg font-semibold">{data.projectStats.total}</p>
                    <p className="text-xs text-muted-foreground">Total</p>
                  </div>
                  <div className="rounded-lg bg-sidebar-accent/50 p-3">
                    <p className="text-lg font-semibold">{data.projectStats.active}</p>
                    <p className="text-xs text-muted-foreground">Active</p>
                  </div>
                  <div className="rounded-lg bg-sidebar-accent/50 p-3">
                    <p className="text-lg font-semibold">{data.projectStats.migrations}</p>
                    <p className="text-xs text-muted-foreground">Migrations</p>
                  </div>
                  <div className="rounded-lg bg-sidebar-accent/50 p-3">
                    <p className="text-lg font-semibold">{data.projectStats.drift}</p>
                    <p className="text-xs text-amber-500">Drift Events</p>
                  </div>
                </div>
                <div className="px-4 py-3 text-xs font-medium text-muted-foreground uppercase tracking-wider border-t">
                  Recent Projects
                </div>
                {data.projectsList.map((project) => (
                  <a
                    key={project.name}
                    href="/projects"
                    onClick={(e) => { e.preventDefault(); router.push("/projects"); }}
                    className="flex flex-col gap-1 border-b p-4 text-sm leading-tight last:border-b-0 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
                  >
                    <div className="font-medium">{project.name}</div>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                      <span className={`inline-block size-1.5 rounded-full ${project.status === "active" ? "bg-green-500" : project.status === "inProgress" ? "bg-yellow-500" : "bg-red-500"}`} />
                      <span>{project.updated}</span>
                    </div>
                  </a>
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
                {data.schemas.map((schema) => (
                  <div key={schema.name} className="flex items-center justify-between border-b p-4 text-sm last:border-b-0 hover:bg-sidebar-accent">
                    <div className="flex items-center gap-3">
                      <Database className="size-4 text-muted-foreground" />
                      <div>
                        <div className="font-medium">{schema.name}</div>
                        <div className="text-xs text-muted-foreground">{schema.tables} tables</div>
                      </div>
                    </div>
                    <span className="text-xs text-muted-foreground">{schema.lastUpdated}</span>
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
                {data.teamMembers.map((member) => (
                  <div key={member.name} className="flex items-center gap-3 border-b p-4 text-sm last:border-b-0 hover:bg-sidebar-accent">
                    <Avatar className="size-8">
                      <AvatarFallback className="text-xs">{member.initials}</AvatarFallback>
                    </Avatar>
                    <div className="min-w-0">
                      <div className="font-medium truncate">{member.name}</div>
                      <div className="text-xs text-muted-foreground">{member.role}</div>
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
                <a href="#" className="flex items-center gap-3 border-b p-4 text-sm hover:bg-sidebar-accent">
                  <Settings className="size-4 text-muted-foreground" />
                  <span>Profile Settings</span>
                </a>
                <a href="#" className="flex items-center gap-3 border-b p-4 text-sm hover:bg-sidebar-accent">
                  <Database className="size-4 text-muted-foreground" />
                  <span>Database Connections</span>
                </a>
                <a href="#" className="flex items-center gap-3 border-b p-4 text-sm hover:bg-sidebar-accent">
                  <Users className="size-4 text-muted-foreground" />
                  <span>Team Management</span>
                </a>
                <a href="#" className="flex items-center gap-3 border-b p-4 text-sm hover:bg-sidebar-accent">
                  <Bell className="size-4 text-muted-foreground" />
                  <span>Notifications</span>
                </a>
              </SidebarGroupContent>
            </SidebarGroup>
          )}
        </SidebarContent>
      </Sidebar>
    </Sidebar>
  )
}

