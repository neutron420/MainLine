"use client";

import React, { useState, useRef } from "react";
import { useRouter } from "next/navigation";
import Image from "next/image";

export default function OtpPage() {
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
    // TODO: verify OTP via backend — otp.join("")
    router.push("/forgot-password/reset");
  };

  const handleResend = () => {
    // TODO: call ForgotPassword RPC again
  };

  return (
    <div suppressHydrationWarning className="min-h-screen flex items-center justify-center bg-gradient-to-br from-gray-50 to-gray-100 p-4">
      <div className="w-full max-w-md bg-white rounded-2xl shadow-xl overflow-hidden border border-gray-100 relative">
        <div className="absolute top-0 left-0 right-0 h-48 bg-gradient-to-b from-blue-100 via-blue-50 to-transparent opacity-40 blur-3xl -mt-20"></div>
        <div className="p-8">
          <div className="flex flex-col items-center mb-8">
            <div className="bg-white p-3 rounded-2xl shadow-lg mb-6 flex items-center justify-center">
              <Image src="/logo.png" alt="SchemaHub" width={56} height={56} className="object-contain" />
            </div>
            <div className="p-0">
              <h2 className="text-2xl font-bold text-gray-900 text-center">Verify OTP</h2>
              <p className="text-center text-gray-500 mt-2">Enter the 6-digit code sent to your email</p>
            </div>
          </div>

          <form onSubmit={handleSubmit} className="space-y-6 p-0">
            <div className="flex justify-center gap-3">
              {otp.map((digit, index) => (
                <input
                  key={index}
                  ref={(el) => { inputRefs.current[index] = el; }}
                  type="text"
                  inputMode="numeric"
                  maxLength={1}
                  value={digit}
                  onChange={(e) => handleChange(index, e.target.value)}
                  onKeyDown={(e) => handleKeyDown(index, e)}
                  className="w-14 h-16 text-center text-xl font-bold bg-white border-2 border-gray-300 text-gray-900 rounded-xl shadow-sm focus-visible:ring-2 focus-visible:ring-blue-500 focus:border-blue-500 ring-offset-background focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50 transition-all"
                />
              ))}
            </div>

            <button type="submit" className="w-full h-12 bg-gradient-to-t from-blue-600 via-blue-500 to-blue-400 hover:from-blue-700 hover:via-blue-600 hover:to-blue-500 text-white font-medium rounded-lg transition-all duration-200 shadow-sm hover:shadow-md hover:shadow-blue-100 active:scale-[0.98] inline-flex items-center justify-center whitespace-nowrap text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50">
              Verify OTP
            </button>
          </form>

          <div className="p-0 mt-6 text-center">
            <p className="text-sm text-gray-500">
              Didn&apos;t receive the code?{" "}
              <button onClick={handleResend} className="text-blue-600 hover:underline font-medium bg-transparent border-none cursor-pointer text-sm">
                Resend
              </button>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}


