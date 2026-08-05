"use client";

import { useEffect, useRef, useState } from "react";
import { useParams, useRouter, useSearchParams } from "next/navigation";

import { Background } from "@/components/background";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { authClient } from "@/lib/api/clients";
import { authStore } from "@/lib/api/auth-store";
import { getApiErrorMessage } from "@/lib/api/errors";
import { consumePkceVerifier } from "@/lib/api/pkce";

export default function OAuthCallbackPage() {
  const params = useParams<{ provider: string }>();
  const searchParams = useSearchParams();
  const router = useRouter();
  const handled = useRef(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (handled.current) return;
    handled.current = true;

    const provider = params.provider;
    const code = searchParams.get("code");
    const state = searchParams.get("state");
    const codeVerifier = consumePkceVerifier();

    if (!code || !state) {
      router.push("/login?error=invalid_oauth_response");
      return;
    }

    const exchange = async () => {
      try {
        const res = await authClient.handleOAuthCallback({
          provider,
          code,
          state,
          codeVerifier,
        });
        authStore.setTokens(res.accessToken, res.refreshToken);
        router.push("/dashboard");
      } catch (err) {
        setError(getApiErrorMessage(err));
      }
    };

    void exchange();
  }, [params.provider, searchParams, router]);

  return (
    <Background>
      <section className="py-28 lg:pt-44 lg:pb-32">
        <div className="container">
          <div className="flex flex-col gap-4">
            <Card className="mx-auto w-full max-w-sm">
              <CardHeader className="flex flex-col items-center space-y-0">
                <p className="mb-2 text-2xl font-bold">
                  {error ? "Sign in failed" : "Completing sign in"}
                </p>
                <p className="text-muted-foreground text-center">
                  {error
                    ? error
                    : "Please wait while we authenticate you."}
                </p>
              </CardHeader>
              <CardContent className="flex flex-col items-center gap-4 pb-8">
                {!error && (
                  <div className="border-primary h-10 w-10 animate-spin rounded-full border-4 border-t-transparent" />
                )}
                {error && (
                  <Button onClick={() => router.push("/login")}>Back to login</Button>
                )}
              </CardContent>
            </Card>
          </div>
        </div>
      </section>
    </Background>
  );
}
