"use client";

import { useState } from "react";

import Image from "next/image";
import Link from "next/link";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Background } from "@/components/background";
import { useResetPassword } from "@/lib/api/hooks/use-auth";
import { getApiErrorMessage } from "@/lib/api/errors";
import { CheckCircle2 } from "lucide-react";

const ResetPassword = () => {
  const resetPassword = useResetPassword();
  const [token, setToken] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [done, setDone] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    if (password.length < 6) {
      setError("Password must be at least 6 characters");
      return;
    }
    if (password !== confirm) {
      setError("Passwords do not match");
      return;
    }
    try {
      await resetPassword.mutateAsync({ token: token.trim(), password });
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
                    <p className="mb-2 text-2xl font-bold">Password reset</p>
                    <p className="text-muted-foreground text-center">
                      Your password has been updated. You can now log in.
                    </p>
                  </>
                ) : (
                  <>
                    <p className="mb-2 text-2xl font-bold">Set a new password</p>
                    <p className="text-muted-foreground text-center">
                      Enter the token from your email and choose a new password.
                    </p>
                  </>
                )}
              </CardHeader>
              {!done && (
                <CardContent>
                  <form onSubmit={submit} className="grid gap-4">
                    <div className="grid gap-2">
                      <Label htmlFor="token">Reset Token</Label>
                      <Input
                        id="token"
                        placeholder="Paste the token from your email"
                        required
                        value={token}
                        onChange={(e) => setToken(e.target.value)}
                      />
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="password">New Password</Label>
                      <Input
                        id="password"
                        type="password"
                        placeholder="At least 6 characters"
                        required
                        autoComplete="new-password"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                      />
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="confirm">Confirm Password</Label>
                      <Input
                        id="confirm"
                        type="password"
                        placeholder="Repeat your new password"
                        required
                        autoComplete="new-password"
                        value={confirm}
                        onChange={(e) => setConfirm(e.target.value)}
                      />
                    </div>
                    {error && <p className="text-destructive text-sm">{error}</p>}
                    <Button type="submit" className="mt-2 w-full" disabled={resetPassword.isPending}>
                      {resetPassword.isPending ? "Resetting..." : "Reset password"}
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

export default ResetPassword;
