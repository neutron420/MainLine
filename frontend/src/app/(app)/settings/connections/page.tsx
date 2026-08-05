"use client";

import { useState } from "react";
import Link from "next/link";
import { Search, Link2, Database, Loader2 } from "lucide-react";
import { FcGoogle } from "react-icons/fc";
import { FaGithub, FaSlack } from "react-icons/fa";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { NotificationsPopover } from "@/components/notifications-popover";
import { Tooltip } from "@heroui/react";
import { ProjectConnectionsCard } from "@/components/project-connections-card";
import {
  useGetOAuthUrl,
  useListLinkedIdentities,
  useUnlinkOAuthIdentity,
} from "@/lib/api/hooks/use-auth";
import { useProjects } from "@/lib/api/hooks/use-projects";

type Provider = "github" | "google" | "slack";

const providers: { id: Provider; name: string; icon: React.ComponentType<{ className?: string }> }[] = [
  { id: "github", name: "GitHub", icon: FaGithub },
  { id: "google", name: "Google", icon: FcGoogle },
  { id: "slack", name: "Slack", icon: FaSlack },
];

export default function LinkedAccountsPage() {
  const { data: projects } = useProjects();
  const { data: identities } = useListLinkedIdentities();
  const unlinkIdentity = useUnlinkOAuthIdentity();
  const getOAuthUrl = useGetOAuthUrl();
  const [linking, setLinking] = useState<Provider | null>(null);

  const linkAccount = async (provider: Provider) => {
    setLinking(provider);
    try {
      const url = await getOAuthUrl.mutateAsync({
        provider,
        redirectTo: window.location.origin,
        linking: true,
      });
      if (url) window.open(url, "_blank", "noopener,noreferrer");
    } finally {
      setLinking(null);
    }
  };

  return (
            <div className="flex flex-1 flex-col gap-6 p-6">
        <div className="flex flex-wrap items-center gap-3">
            <div className="relative">
              <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground" />
              <Input
                placeholder="Search..."
                className="w-[180px] lg:w-[220px] h-9 pl-8 text-sm"
              />
            </div>
          </div>

        <div className="flex flex-1 flex-col gap-6 p-6 max-w-2xl w-full mx-auto">
          <div>
            <div className="flex items-center gap-3 flex-wrap">
              <h1 className="text-2xl font-semibold tracking-tight flex items-center gap-2">
                <Link2 className="size-6" />
                Linked Accounts
              </h1>
            </div>
            <p className="text-sm text-muted-foreground mt-1">Connect accounts and manage database connections</p>
          </div>

          <Card>
            <CardHeader className="border-0">
              <CardTitle className="text-base">Account connections</CardTitle>
            </CardHeader>
            <CardContent className="pt-0">
              <div className="divide-y">
                {providers.map((provider) => {
                  const identity = (identities ?? []).find((i) => i.provider === provider.id);
                  const isLinked = Boolean(identity);
                  const isLinking = linking === provider.id;
                  return (
                    <div key={provider.id} className="flex items-center gap-4 py-4">
                      <div className="flex size-10 shrink-0 items-center justify-center rounded-lg bg-muted">
                        <provider.icon className="size-5 text-muted-foreground" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium">{provider.name}</p>
                        <p className="text-xs text-muted-foreground truncate mt-0.5">
                          {isLinked ? identity?.providerEmail : "Not linked"}
                        </p>
                      </div>
                      {isLinked ? (
                        <>
                          <Badge variant="default" className="text-[10px] px-1.5 py-0 gap-1">
                            <span className="size-1.5 rounded-full bg-emerald-500" />
                            Connected
                          </Badge>
                          <Button
                            variant="outline"
                            size="sm"
                            className="h-8 text-xs"
                            disabled={unlinkIdentity.isPending}
                            onClick={() => unlinkIdentity.mutate(provider.id)}
                          >
                            Unlink
                          </Button>
                        </>
                      ) : (
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-8 text-xs gap-1.5"
                          disabled={isLinking}
                          onClick={() => linkAccount(provider.id)}
                        >
                          {isLinking && <Loader2 className="size-3.5 animate-spin" />}
                          {isLinking ? "Opening..." : "Link"}
                        </Button>
                      )}
                    </div>
                  );
                })}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="border-0">
              <CardTitle className="text-base flex items-center gap-2">
                <Database className="size-4" />
                Connected databases
              </CardTitle>
            </CardHeader>
            <CardContent className="pt-0">
              <div className="space-y-6">
                {projects?.map((project) => (
                  <ProjectConnectionsCard key={project.id} project={project} />
                ))}
                {(!projects || projects.length === 0) && (
                  <p className="text-sm text-muted-foreground">No projects yet.</p>
                )}
              </div>
              <Separator className="my-4" />
              <Link href="/projects">
                <Button variant="outline" size="sm" className="h-9">
                  Add Connection
                </Button>
              </Link>
            </CardContent>
          </Card>
        </div>
      </div>
  );
}
