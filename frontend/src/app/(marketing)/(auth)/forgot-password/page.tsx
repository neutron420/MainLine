"use client";

import { useState } from "react";

import Image from "next/image";
import Link from "next/link";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Background } from "@/components/background";
import { useForgotPassword } from "@/lib/api/hooks/use-auth";
import { getApiErrorMessage } from "@/lib/api/errors";
import { CheckCircle2 } from "lucide-react";

const ForgotPassword = () => {
  const forgotPassword = useForgotPassword();
  const [email, setEmail] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [sent, setSent] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    try {
      await forgotPassword.mutateAsync(email);
      setSent(true);
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
                {sent ? (
                  <>
                    <CheckCircle2 className="size-10 text-emerald-500 mb-3" />
                    <p className="mb-2 text-2xl font-bold">Check your email</p>
                    <p className="text-muted-foreground text-center">
                      We sent a password reset link to <span className="font-medium">{email}</span>.
                    </p>
                  </>
                ) : (
                  <>
                    <p className="mb-2 text-2xl font-bold">Forgot password?</p>
                    <p className="text-muted-foreground text-center">
                      Enter your email and we&apos;ll send you a reset link.
                    </p>
                  </>
                )}
              </CardHeader>
              {!sent && (
                <CardContent>
                  <form onSubmit={submit} className="grid gap-4">
                    <div className="grid gap-2">
                      <Label htmlFor="email">Email</Label>
                      <Input
                        id="email"
                        type="email"
                        placeholder="Enter your email"
                        required
                        autoComplete="email"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                      />
                    </div>
                    {error && <p className="text-destructive text-sm">{error}</p>}
                    <Button type="submit" className="mt-2 w-full" disabled={forgotPassword.isPending}>
                      {forgotPassword.isPending ? "Sending..." : "Send reset link"}
                    </Button>
                    <div className="text-center">
                      <Link href="/login" className="text-primary text-sm font-medium">
                        Back to login
                      </Link>
                    </div>
                  </form>
                </CardContent>
              )}
              {sent && (
                <CardContent className="grid gap-4">
                  <div className="text-center">
                    <Link href="/login" className="text-primary text-sm font-medium">
                      Back to login
                    </Link>
                  </div>
                </CardContent>
              )}
            </Card>
          </div>
        </div>
      </section>
    </Background>
  );
};

export default ForgotPassword;
