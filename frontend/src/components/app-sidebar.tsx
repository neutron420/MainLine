"use client"

import * as React from "react"
import { Database, FolderKanban, GitPullRequest, GanttChartSquare, LayoutDashboard, Settings, Users } from "lucide-react"

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

const data = {
  user: {
    name: "R.K Singh",
    email: "rk@mainline.dev",
    avatar: "/avatars/default.jpg",
  },
  navMain: [
    { title: "Dashboard", url: "/dashboard", icon: LayoutDashboard, isActive: true },
    { title: "Projects", url: "/projects", icon: FolderKanban, isActive: false },
    { title: "Reviews", url: "/reviews", icon: GitPullRequest, isActive: false },
    { title: "Schemas", url: "/schemas", icon: Database, isActive: false },
    { title: "Team", url: "/team", icon: Users, isActive: false },
    { title: "Settings", url: "/settings", icon: Settings, isActive: false },
  ],
  projects: [
    { name: "User Service Schema", updated: "2 hours ago", status: "active" as const },
    { name: "Payment DB Migration", updated: "5 hours ago", status: "inProgress" as const },
    { name: "Analytics Warehouse", updated: "1 day ago", status: "active" as const },
    { name: "Legacy CRM Schema", updated: "3 days ago", status: "onHold" as const },
  ],
  reviews: [
    { title: "Add users table index", project: "User Service", author: "Alice", date: "1 hour ago" },
    { title: "Update payment status enum", project: "Payment DB", author: "Bob", date: "3 hours ago" },
    { title: "Drop legacy column", project: "CRM", author: "Charlie", date: "1 day ago" },
  ],
}

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const [activeItem, setActiveItem] = React.useState(data.navMain[0])
  const { setOpen } = useSidebar()

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
              <SidebarMenuButton size="lg" asChild className="md:h-8 md:p-0">
                <a href="/dashboard">
                  <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                    <GanttChartSquare className="size-4" />
                  </div>
                  <div className="grid flex-1 text-left text-sm leading-tight">
                    <span className="truncate font-medium">Mainline</span>
                    <span className="truncate text-xs">Schema Hub</span>
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
                        setOpen(true)
                      }}
                      isActive={activeItem?.title === item.title}
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
              {activeItem?.title === "Dashboard" ? "Overview" : activeItem?.title}
            </div>
          </div>
        </SidebarHeader>
        <SidebarContent>
          {activeItem?.title === "Dashboard" && (
            <SidebarGroup className="px-0">
              <SidebarGroupContent>
                <div className="px-4 py-3 text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Recent Projects
                </div>
                {data.projects.map((project) => (
                  <a
                    key={project.name}
                    href="#"
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

          {activeItem?.title === "Reviews" && (
            <SidebarGroup className="px-0">
              <SidebarGroupContent>
                <div className="px-4 py-3 text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Pending Reviews
                </div>
                {data.reviews.map((review) => (
                  <a
                    key={review.title}
                    href="#"
                    className="flex flex-col gap-1 border-b p-4 text-sm leading-tight last:border-b-0 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
                  >
                    <div className="font-medium">{review.title}</div>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                      <span>{review.project}</span>
                      <span>·</span>
                      <span>{review.author}</span>
                      <span className="ml-auto">{review.date}</span>
                    </div>
                  </a>
                ))}
              </SidebarGroupContent>
            </SidebarGroup>
          )}

          {(activeItem?.title !== "Dashboard" && activeItem?.title !== "Reviews") && (
            <div className="flex items-center justify-center h-full p-8 text-sm text-muted-foreground">
              Select a section from the sidebar
            </div>
          )}
        </SidebarContent>
      </Sidebar>
    </Sidebar>
  )
}
