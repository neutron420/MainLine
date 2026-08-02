"use client";

import { Suspense, useState } from "react";

import Image from "next/image";
import Link from "next/link";
import { useSearchParams } from "next/navigation";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Background } from "@/components/background";
import { useVerifyEmail } from "@/lib/api/hooks/use-auth";
import { getApiErrorMessage } from "@/lib/api/errors";
import { CheckCircle2 } from "lucide-react";

const VerifyEmailContent = () => {
  const searchParams = useSearchParams();
  const verifyEmail = useVerifyEmail();
  const [token, setToken] = useState(searchParams.get("token") ?? "");
  const [error, setError] = useState<string | null>(null);
  const [done, setDone] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    if (!token.trim()) {
      setError("Verification token is required");
      return;
    }
    try {
      await verifyEmail.mutateAsync(token.trim());
      setDone(true);
    } catch (err) {
      setError(getApiErrorMessage(err));
    }
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
                {done ? (
                  <>
                    <CheckCircle2 className="size-10 text-emerald-500 mb-3" />
                    <p className="mb-2 text-2xl font-bold">Email verified</p>
                    <p className="text-muted-foreground text-center">
                      Your email has been verified. You can now log in.
                    </p>
                  </>
                ) : (
                  <>
                    <p className="mb-2 text-2xl font-bold">Verify your email</p>
                    <p className="text-muted-foreground text-center">
                      Click the link we sent you, or paste the token from your email.
                    </p>
                  </>
                )}
              </CardHeader>
              {!done && (
                <CardContent>
                  <form onSubmit={submit} className="grid gap-4">
                    <div className="grid gap-2">
                      <Label htmlFor="token">Verification Token</Label>
                      <Input
                        id="token"
                        placeholder="Paste the token from your email"
                        required
                        value={token}
                        onChange={(e) => setToken(e.target.value)}
                      />
                    </div>
                    {error && <p className="text-destructive text-sm">{error}</p>}
                    <Button type="submit" className="mt-2 w-full" disabled={verifyEmail.isPending}>
                      {verifyEmail.isPending ? "Verifying..." : "Verify email"}
                    </Button>
                    <div className="text-center">
                      <Link href="/login" className="text-primary text-sm font-medium">
                        Back to login
                      </Link>
                    </div>
                  </form>
                </CardContent>
              )}
              {done && (
                <CardContent className="grid gap-4">
                  <Link href="/login">
                    <Button className="w-full">Log in</Button>
                  </Link>
                </CardContent>
              )}
            </Card>
          </div>
        </div>
      </section>
    </Background>
  );
};

export default function VerifyEmailPage() {
  return (
    <Suspense fallback={null}>
      <VerifyEmailContent />
    </Suspense>
  );
}
