"use client";

import { useEffect, useRef } from "react";
import { useSearchParams, useRouter } from "next/navigation";

import { Background } from "@/components/background";
import { Card, CardContent, CardHeader } from "@/components/ui/card";

export default function OAuthCallbackPage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const handled = useRef(false);

  useEffect(() => {
    if (handled.current) return;
    handled.current = true;

    const code = searchParams.get("code");
    const state = searchParams.get("state");

    if (!code || !state) {
      router.push("/login?error=invalid_oauth_response");
      return;
    }

    router.push("/login");
  }, [searchParams, router]);

  return (
    <Background>
      <section className="py-28 lg:pt-44 lg:pb-32">
        <div className="container">
          <div className="flex flex-col gap-4">
            <Card className="mx-auto w-full max-w-sm">
              <CardHeader className="flex flex-col items-center space-y-0">
                <p className="mb-2 text-2xl font-bold">Completing sign in</p>
                <p className="text-muted-foreground text-center">
                  Please wait while we authenticate you.
                </p>
              </CardHeader>
              <CardContent className="flex justify-center pb-8">
                <div className="border-primary h-10 w-10 animate-spin rounded-full border-4 border-t-transparent" />
              </CardContent>
            </Card>
          </div>
        </div>
      </section>
    </Background>
  );
}
