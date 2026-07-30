"use client"

import { Bell, CheckCircle2, Info, AlertTriangle, AlertCircle, X } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { Separator } from "@/components/ui/separator"
const notifications = [
  {
    id: "1",
    icon: CheckCircle2,
    title: "Migration successful",
    description: "User Service Schema v1.2.0 deployed to production successfully.",
    color: "text-green-500",
    time: "2 hours ago",
  },
  {
    id: "2",
    icon: Info,
    title: "New review available",
    description: "Alice requested your review on 'Add users table index'.",
    color: "text-blue-500",
    time: "3 hours ago",
  },
  {
    id: "3",
    icon: AlertTriangle,
    title: "Subscription expiring",
    description: "Your Pro plan renews in 3 days. Update billing to avoid interruption.",
    color: "text-amber-500",
    time: "1 day ago",
  },
  {
    id: "4",
    icon: AlertCircle,
    title: "Schema drift detected",
    description: "Analytics Warehouse schema differs from staging environment.",
    color: "text-red-500",
    time: "2 days ago",
  },
]

export function NotificationsPopover() {
  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="icon" className="relative size-11">
          <Bell className="size-5" />
          <span className="absolute -top-0.5 -right-0.5 flex size-4 items-center justify-center rounded-full bg-red-500 text-[9px] font-medium text-white">
            {notifications.length}
          </span>
        </Button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-[380px] p-0">
        <div className="flex items-center justify-between px-4 py-3">
          <span className="text-sm font-semibold">Notifications</span>
          <Button variant="ghost" size="icon" className="size-7">
            <X className="size-3.5" />
          </Button>
        </div>
        <Separator />
        <div className="max-h-[380px] overflow-y-auto p-3 space-y-2">
          {notifications.map((n) => (
            <Alert key={n.id} className="cursor-pointer transition-colors hover:bg-accent">
              <n.icon className={n.color} />
              <AlertTitle>{n.title}</AlertTitle>
              <AlertDescription>
                {n.description}
                <span className="block text-[11px] text-muted-foreground mt-1">{n.time}</span>
              </AlertDescription>
            </Alert>
          ))}
        </div>
        <Separator />
        <div className="p-2">
          <Button variant="ghost" size="sm" className="w-full text-xs h-9">
            View all notifications
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  )
}
