"use client"

import { useMemo, useState } from "react"
import { Bell, CheckCircle2, Info, AlertTriangle, AlertCircle, X, Database, GitBranch, Table2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { Separator } from "@/components/ui/separator"
import { useAuditEntries } from "@/lib/api/hooks/use-audit"
import { useEventStream } from "@/lib/api/hooks/use-realtime"
import type { SchemaEvent } from "@/lib/gen/event/v1/event_messages_pb"

interface NotificationItem {
  id: string
  icon: React.ComponentType<{ className?: string }>
  title: string
  description: string
  color: string
  time: string
}

function formatTime(value: string): string {
  if (!value) return ""
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  const diffMs = Date.now() - date.getTime()
  const minutes = Math.floor(diffMs / 60_000)
  if (minutes < 1) return "just now"
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  if (days < 7) return `${days}d ago`
  return date.toLocaleDateString()
}

function iconFor(keyword: string): { icon: React.ComponentType<{ className?: string }>; color: string } {
  if (/migration|rollback/i.test(keyword)) return { icon: GitBranch, color: "text-blue-500" }
  if (/drift/i.test(keyword)) return { icon: AlertTriangle, color: "text-amber-500" }
  if (/connection|database/i.test(keyword)) return { icon: Database, color: "text-violet-500" }
  if (/schema|table|column|index/i.test(keyword)) return { icon: Table2, color: "text-emerald-500" }
  if (/fail|error|delete|remove/i.test(keyword)) return { icon: AlertCircle, color: "text-red-500" }
  if (/complete|success|create|update|approved/i.test(keyword)) return { icon: CheckCircle2, color: "text-green-500" }
  return { icon: Info, color: "text-blue-500" }
}

function eventToNotification(event: SchemaEvent): NotificationItem {
  const { icon, color } = iconFor(event.type)
  const title = event.type
    .replace(/[._]/g, " ")
    .replace(/\b\w/g, (c) => c.toUpperCase())
  const actor = event.actor?.email ? ` · ${event.actor.email}` : ""
  return {
    id: event.id,
    icon,
    title,
    description: `${event.resource?.type ?? "schema"}${actor}`,
    color,
    time: formatTime(event.timestamp),
  }
}

function auditToNotification(entry: { id: string; eventType: string; action: string; actorEmail: string; createdAt: string }): NotificationItem {
  const { icon, color } = iconFor(`${entry.eventType} ${entry.action}`)
  return {
    id: entry.id,
    icon,
    title: entry.action || entry.eventType,
    description: entry.actorEmail
      ? `${entry.actorEmail} · ${entry.eventType}`
      : entry.eventType,
    color,
    time: formatTime(entry.createdAt),
  }
}

export function NotificationsPopover() {
  const { data: auditEntries } = useAuditEntries()
  const [open, setOpen] = useState(false)
  const { events: liveEvents, connected } = useEventStream({
    maxEvents: 50,
    enabled: open,
  })
  const [dismissed, setDismissed] = useState<Set<string>>(new Set())

  const notifications = useMemo<NotificationItem[]>(() => {
    const live = liveEvents.map(eventToNotification)
    const history = (auditEntries ?? []).map(auditToNotification)
    const merged = new Map<string, NotificationItem>()
    for (const item of [...live, ...history]) {
      if (!merged.has(item.id)) merged.set(item.id, item)
    }
    return Array.from(merged.values())
      .filter((n) => !dismissed.has(n.id))
      .slice(0, 30)
  }, [liveEvents, auditEntries, dismissed])

  const unread = liveEvents.filter((e) => !dismissed.has(e.id)).length

  const clearAll = () => {
    setDismissed((prev) => {
      const next = new Set(prev)
      for (const n of notifications) next.add(n.id)
      return next
    })
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="icon" className="relative size-11">
          <Bell className="size-5" />
          {unread > 0 && (
            <span className="absolute -top-0.5 -right-0.5 flex size-4 items-center justify-center rounded-full bg-red-500 text-[9px] font-medium text-white">
              {unread > 9 ? "9+" : unread}
            </span>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-[380px] p-0">
        <div className="flex items-center justify-between px-4 py-3">
          <span className="text-sm font-semibold flex items-center gap-2">
            Notifications
            <span className={`inline-flex items-center gap-1 text-[10px] font-medium ${connected ? "text-emerald-600" : "text-muted-foreground"}`}>
              <span className={`size-1.5 rounded-full ${connected ? "bg-emerald-500" : "bg-muted-foreground"}`} />
              {connected ? "LIVE" : "OFFLINE"}
            </span>
          </span>
          <div className="flex items-center gap-1">
            <Button variant="ghost" size="sm" className="h-7 text-xs" onClick={clearAll}>
              Clear
            </Button>
            <Button variant="ghost" size="icon" className="size-7">
              <X className="size-3.5" />
            </Button>
          </div>
        </div>
        <Separator />
        <div className="max-h-[380px] overflow-y-auto p-3 space-y-2">
          {notifications.length === 0 && (
            <p className="text-sm text-muted-foreground text-center py-8">
              No notifications yet
            </p>
          )}
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
      </PopoverContent>
    </Popover>
  )
}
