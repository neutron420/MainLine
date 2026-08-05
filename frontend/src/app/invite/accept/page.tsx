"use client";

import { Suspense, useEffect, useState } from "react";

import Image from "next/image";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";

import { Background } from "@/components/background";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { useAcceptInvitation } from "@/lib/api/hooks/use-projects";
import { useAuth } from "@/lib/api/auth-provider";
import { getApiErrorMessage } from "@/lib/api/errors";
import { Loader2, CheckCircle2, XCircle } from "lucide-react";

const InviteAcceptContent = () => {
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get("token") ?? "";
  const { isAuthenticated, isLoading } = useAuth();
  const accept = useAcceptInvitation();
  const [acceptedName, setAcceptedName] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isLoading || acceptedName || error) return;
    if (!token) {
      setError("This invitation link is invalid. It may be missing its token.");
      return;
    }
    if (!isAuthenticated) {
      router.push(`/login?redirect=${encodeURIComponent(`/invite/accept?token=${token}`)}`);
      return;
    }
    accept.mutate(token, {
      onSuccess: (res) => {
        setAcceptedName(res.projectName || "");
      },
      onError: (err) => {
        setError(getApiErrorMessage(err));
      },
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isLoading, isAuthenticated, token]);

  const renderBody = () => {
    if (isLoading) {
      return (
        <div className="flex flex-col items-center gap-3">
          <Loader2 className="size-8 animate-spin text-muted-foreground" />
          <p className="text-sm text-muted-foreground">Checking invitation…</p>
        </div>
      );
    }
    if (acceptedName !== null) {
      return (
        <div className="flex flex-col items-center gap-3 text-center">
          <CheckCircle2 className="size-10 text-emerald-500" />
          <p className="text-xl font-bold">You&apos;re in!</p>
          <p className="text-sm text-muted-foreground">
            You joined <strong>{acceptedName}</strong>. Take a look around.
          </p>
          <Button
            className="w-full mt-2"
            onClick={() =>
              router.push(`/projects/${accept.data?.projectId ?? ""}`)
            }
          >
            Open project
          </Button>
        </div>
      );
    }
    return (
      <div className="flex flex-col items-center gap-3 text-center">
        <XCircle className="size-10 text-red-500" />
        <p className="text-xl font-bold">Invitation failed</p>
        <p className="text-sm text-muted-foreground">{error ?? "Unknown error"}</p>
        <Link href="/dashboard" className="w-full">
          <Button variant="outline" className="w-full">
            Go to dashboard
          </Button>
        </Link>
      </div>
    );
  };

  return (
    <Background>
      <section className="py-28 lg:pt-44 lg:pb-32">
        <div className="container">
          <div className="flex flex-col gap-4">
            <Card className="mx-auto w-full max-w-sm">
              <CardHeader className="flex flex-col items-center space-y-0">
                <Image
                  src="/logo.svg"
                  alt="logo"
                  width={94}
                  height={18}
                  className="mb-7 dark:invert"
                />
              </CardHeader>
              <CardContent className="grid gap-4">{renderBody()}</CardContent>
            </Card>
          </div>
        </div>
      </section>
    </Background>
  );
};

export default function InviteAcceptPage() {
  return (
    <Suspense fallback={null}>
      <InviteAcceptContent />
    </Suspense>
  );
}
