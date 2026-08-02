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
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useRegister } from "@/lib/api/hooks/use-auth";
import { getApiErrorMessage } from "@/lib/api/errors";
import { authClient } from "@/lib/api/clients";

const signupSchema = z.object({
  name: z.string().min(2, "Enter your name"),
  email: z.string().email("Enter a valid email"),
  password: z.string().min(8, "Password must be at least 8 characters"),
});

type SignupForm = z.infer<typeof signupSchema>;

const Signup = () => {
  const router = useRouter();
  const registerMutation = useRegister();
  const [error, setError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<SignupForm>({
    resolver: zodResolver(signupSchema),
    defaultValues: { name: "", email: "", password: "" },
  });

  const onSubmit = handleSubmit(async (values) => {
    setError(null);
    try {
      await registerMutation.mutateAsync({
        displayName: values.name,
        email: values.email,
        password: values.password,
      });
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
      setError("Could not start OAuth sign-up. Please try email.");
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
                <p className="mb-2 text-2xl font-bold">Start your free trial</p>
                <p className="text-muted-foreground">
                  Sign up in less than 2 minutes.
                </p>
              </CardHeader>
              <CardContent>
                <form onSubmit={onSubmit} className="grid gap-4">
                  <div className="grid gap-2">
                    <Label htmlFor="name">Name</Label>
                    <Input
                      id="name"
                      type="text"
                      placeholder="Enter your name"
                      required
                      autoComplete="name"
                      {...register("name")}
                    />
                    {errors.name && (
                      <p className="text-destructive text-sm">{errors.name.message}</p>
                    )}
                  </div>
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
                      autoComplete="new-password"
                      {...register("password")}
                    />
                    <p className="text-muted-foreground mt-1 text-sm">
                      Must be at least 8 characters.
                    </p>
                    {errors.password && (
                      <p className="text-destructive text-sm">{errors.password.message}</p>
                    )}
                  </div>
                  {error && (
                    <p className="text-destructive text-sm">{error}</p>
                  )}
                  <Button type="submit" className="mt-2 w-full" disabled={isSubmitting}>
                    {isSubmitting ? "Creating account..." : "Create an account"}
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
                  <p>Already have an account?</p>
                  <Link href="/login" className="text-primary font-medium">
                    Log in
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

export default Signup;
