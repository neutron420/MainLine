"use client";

import { useState } from "react";

import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { FcGoogle } from "react-icons/fc";
import { FaGithub, FaSlack } from "react-icons/fa";
import { z } from "zod";

import { Background } from "@/components/background";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useLogin } from "@/lib/api/hooks/use-auth";
import { getApiErrorMessage } from "@/lib/api/errors";
import { authClient } from "@/lib/api/clients";

const loginSchema = z.object({
  email: z.string().email("Enter a valid email"),
  password: z.string().min(6, "Password must be at least 6 characters"),
});

type LoginForm = z.infer<typeof loginSchema>;

const Login = () => {
  const router = useRouter();
  const loginMutation = useLogin();
  const [error, setError] = useState<string | null>(null);
  const [remember, setRemember] = useState(true);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: "", password: "" },
  });

  const onSubmit = handleSubmit(async (values) => {
    setError(null);
    try {
      await loginMutation.mutateAsync({ ...values, remember });
      router.push("/dashboard");
    } catch (err) {
      setError(getApiErrorMessage(err));
    }
  });

  const startOAuth = async (provider: string) => {
    try {
      const res = await authClient.getOAuthURL({
        provider,
        redirectTo: `${window.location.origin}/auth/callback/${provider}`,
        linking: false,
      });
      window.location.href = res.authUrl;
    } catch {
      setError("Could not start OAuth sign-in. Please try email login.");
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
                <p className="mb-2 text-2xl font-bold">Welcome back</p>
                <p className="text-muted-foreground">
                  Please enter your details.
                </p>
              </CardHeader>
              <CardContent>
                <form onSubmit={onSubmit} className="grid gap-4">
                  <div className="grid gap-2">
                    <Label htmlFor="email">Email</Label>
                    <Input
                      id="email"
                      type="email"
                      placeholder="Enter your email"
                      required
                      autoComplete="email"
                      {...register("email")}
                    />
                    {errors.email && (
                      <p className="text-destructive text-sm">{errors.email.message}</p>
                    )}
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="password">Password</Label>
                    <Input
                      id="password"
                      type="password"
                      placeholder="Enter your password"
                      required
                      autoComplete="current-password"
                      {...register("password")}
                    />
                    {errors.password && (
                      <p className="text-destructive text-sm">{errors.password.message}</p>
                    )}
                  </div>
                  {error && (
                    <p className="text-destructive text-sm">{error}</p>
                  )}
                  <div className="flex justify-between">
                    <div className="flex items-center space-x-2">
                      <Checkbox
                        id="remember"
                        className="border-muted-foreground"
                        checked={remember}
                        onCheckedChange={(checked) => setRemember(checked === true)}
                      />
                      <label
                        htmlFor="remember"
                        className="text-sm leading-none font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                      >
                        Remember me
                      </label>
                    </div>
                    <Link
                      href="/forgot-password"
                      className="text-primary text-sm font-medium"
                    >
                      Forgot password
                    </Link>
                  </div>
                  <Button type="submit" className="mt-2 w-full" disabled={isSubmitting}>
                    {isSubmitting ? "Logging in..." : "Log in"}
                  </Button>
                  <div className="relative my-2">
                    <div className="absolute inset-0 flex items-center">
                      <span className="w-full border-t" />
                    </div>
                    <div className="relative flex justify-center text-xs uppercase">
                      <span className="bg-card px-2 text-muted-foreground">
                        Or continue with
                      </span>
                    </div>
                  </div>
                  <div className="grid grid-cols-3 gap-3">
                    <Button type="button" variant="outline" size="lg" className="w-full" onClick={() => startOAuth("google")}>
                      <FcGoogle className="size-5" />
                    </Button>
                    <Button type="button" variant="outline" className="w-full" onClick={() => startOAuth("github")}>
                      <FaGithub className="size-5" />
                    </Button>
                    <Button type="button" variant="outline" className="w-full" onClick={() => startOAuth("slack")}>
                      <FaSlack className="size-5 text-[#4A154B]" />
                    </Button>
                  </div>
                </form>
                <div className="text-muted-foreground mx-auto mt-8 flex justify-center gap-1 text-sm">
                  <p>Don&apos;t have an account?</p>
                  <Link href="/signup" className="text-primary font-medium">
                    Sign up
                  </Link>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </section>
    </Background>
  );
};

export default Login;
