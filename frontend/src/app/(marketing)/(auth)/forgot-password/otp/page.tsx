"use client";

import { useState, useRef } from "react";
import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";

import { ArrowLeft } from "lucide-react";

import { Background } from "@/components/background";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";

const OtpPage = () => {
  const router = useRouter();
  const [otp, setOtp] = useState(["", "", "", "", "", ""]);
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  const handleChange = (index: number, value: string) => {
    if (value.length > 1) return;
    const newOtp = [...otp];
    newOtp[index] = value;
    setOtp(newOtp);
    if (value && index < 5) {
      inputRefs.current[index + 1]?.focus();
    }
  };

  const handleKeyDown = (index: number, e: React.KeyboardEvent) => {
    if (e.key === "Backspace" && !otp[index] && index > 0) {
      inputRefs.current[index - 1]?.focus();
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    router.push("/forgot-password/reset");
  };

  const handleResend = () => {};

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
                <p className="mb-2 text-2xl font-bold">Check your email</p>
                <p className="text-muted-foreground text-center">
                  Enter the 6-digit code we sent to your email.
                </p>
              </CardHeader>
              <CardContent>
                <form onSubmit={handleSubmit} className="grid gap-6">
                  <div className="grid grid-cols-6 gap-2">
                    {otp.map((digit, index) => (
                      <input
                        key={index}
                        ref={(el) => {
                          inputRefs.current[index] = el;
                        }}
                        type="text"
                        inputMode="numeric"
                        maxLength={1}
                        value={digit}
                        onChange={(e) => handleChange(index, e.target.value)}
                        onKeyDown={(e) => handleKeyDown(index, e)}
                        className="border-input bg-background ring-offset-background focus-visible:ring-ring h-14 w-full rounded-xl border text-center text-xl font-bold shadow-xs transition-all focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-hidden"
                      />
                    ))}
                  </div>
                  <Button type="submit" className="mt-2 w-full">
                    Verify OTP
                  </Button>
                </form>
                <div className="text-muted-foreground mx-auto mt-6 flex flex-col items-center gap-3 text-sm">
                  <p>
                    Didn&apos;t receive the code?{" "}
                    <button
                      onClick={handleResend}
                      className="text-primary cursor-pointer bg-transparent border-none font-medium text-sm px-3 py-2"
                    >
                      Resend
                    </button>
                  </p>
                  <Link
                    href="/forgot-password"
                    className="flex items-center gap-1 text-primary font-medium"
                  >
                    <ArrowLeft className="size-4" />
                    Use a different email
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

export default OtpPage;
