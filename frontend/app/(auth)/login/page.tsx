"use client";

import React, { useState } from "react";
import Link from "next/link";
import Image from "next/image";

interface Errors {
  email?: string;
  password?: string;
}

export default function LoginPage() {
  const [showPassword, setShowPassword] = useState(false);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [errors, setErrors] = useState<Errors>({});
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  const validate = (): Errors => {
    const errs: Errors = {};
    if (!email.trim()) errs.email = "Email is required";
    else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) errs.email = "Enter a valid email address";
    if (!password) errs.password = "Password is required";
    else if (password.length < 8) errs.password = "Password must be at least 8 characters";
    return errs;
  };

  const handleBlur = (field: string) => {
    setTouched((prev) => ({ ...prev, [field]: true }));
    setErrors(validate());
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const errs = validate();
    setErrors(errs);
    setTouched({ email: true, password: true });
    if (Object.keys(errs).length > 0) return;
    // TODO: call Login RPC
  };

  const handleOAuth = (_: string) => {
    // TODO: call GetOAuthURL RPC and redirect
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-gray-50 to-gray-100 p-4">
      <div className="w-full max-w-md bg-white rounded-2xl shadow-xl overflow-hidden border border-gray-100 relative">
        <div className="absolute top-0 left-0 right-0 h-48 bg-gradient-to-b from-blue-100 via-blue-50 to-transparent opacity-40 blur-3xl -mt-20"></div>
        <div className="p-8">
          <div className="flex flex-col items-center mb-8">
            <div className="bg-white p-3 rounded-2xl shadow-lg mb-6 flex items-center justify-center">
              <Image src="/logo.png" alt="SchemaHub" width={56} height={56} className="object-contain" />
            </div>
            <div className="p-0">
              <h2 className="text-2xl font-bold text-gray-900 text-center">Welcome Back</h2>
              <p className="text-center text-gray-500 mt-2">Sign in to continue to your account</p>
            </div>
          </div>

          <form onSubmit={handleSubmit} className="space-y-6 p-0" noValidate>
            <div className="space-y-1">
              <label className="text-sm font-medium text-gray-700">Email</label>
              <input
                value={email}
                onChange={(e) => { setEmail(e.target.value); if (touched.email) setErrors(validate()); }}
                onBlur={() => handleBlur("email")}
                className={`bg-white border-2 shadow-sm transition-all text-gray-900 placeholder:text-gray-400 h-12 rounded-lg focus-visible:ring-2 focus-visible:ring-blue-500 focus:border-blue-500 w-full px-3 py-2 text-sm ring-offset-background focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50 ${errors.email && touched.email ? "border-red-400 focus-visible:ring-red-500 focus:border-red-500" : "border-gray-300"}`}
                placeholder="you@example.com"
              />
              {errors.email && touched.email && <p className="text-red-500 text-xs mt-1">{errors.email}</p>}
            </div>

            <div className="space-y-1">
              <div className="flex justify-between items-center">
                <label className="text-sm font-medium text-gray-700">Password</label>
                <Link href="/forgot-password" className="text-xs text-blue-600 hover:underline">
                  Forgot password?
                </Link>
              </div>
              <div className="relative">
                <input
                  type={showPassword ? "text" : "password"}
                  value={password}
                  onChange={(e) => { setPassword(e.target.value); if (touched.password) setErrors(validate()); }}
                  onBlur={() => handleBlur("password")}
                  className={`bg-white border-2 shadow-sm transition-all text-gray-900 pr-12 h-12 rounded-lg focus-visible:ring-2 focus-visible:ring-blue-500 focus:border-blue-500 w-full px-3 py-2 text-sm ring-offset-background focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50 ${errors.password && touched.password ? "border-red-400 focus-visible:ring-red-500 focus:border-red-500" : "border-gray-300"}`}
                  placeholder="••••••••"
                />
                <button
                  type="button"
                  className="absolute right-1 top-1/2 -translate-y-1/2 text-gray-400 hover:text-blue-600 hover:bg-gray-100 inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 h-9 px-3"
                  onClick={() => setShowPassword(!showPassword)}
                >
                  {showPassword ? "Hide" : "Show"}
                </button>
              </div>
              {errors.password && touched.password && <p className="text-red-500 text-xs mt-1">{errors.password}</p>}
            </div>

            <button type="submit" className="w-full h-12 bg-gradient-to-t from-blue-600 via-blue-500 to-blue-400 hover:from-blue-700 hover:via-blue-600 hover:to-blue-500 text-white font-medium rounded-lg transition-all duration-200 shadow-sm hover:shadow-md hover:shadow-blue-100 active:scale-[0.98] inline-flex items-center justify-center whitespace-nowrap text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50">
              Sign In
            </button>

            <div className="flex items-center my-4">
              <div className="flex-1 h-px bg-gray-200"></div>
              <span className="px-4 text-sm text-gray-400">or continue with</span>
              <div className="flex-1 h-px bg-gray-200"></div>
            </div>

            <div className="grid grid-cols-3 gap-3">
              <button type="button" onClick={() => handleOAuth("google")} className="h-12 bg-white border-2 border-gray-300 shadow-sm transition-all text-gray-700 hover:bg-gray-50 hover:text-blue-600 rounded-lg flex items-center justify-center gap-2 border bg-background inline-flex whitespace-nowrap text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" className="text-gray-700">
                  <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4" />
                  <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853" />
                  <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z" fill="#FBBC05" />
                  <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335" />
                </svg>
                <span className="hidden sm:inline whitespace-nowrap">Google</span>
              </button>

              <button type="button" onClick={() => handleOAuth("github")} className="h-12 bg-white border-2 border-gray-300 shadow-sm transition-all text-gray-700 hover:bg-gray-50 hover:text-black rounded-lg flex items-center justify-center gap-2 border bg-background inline-flex whitespace-nowrap text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" className="text-gray-700">
                  <path d="M12 0C5.374 0 0 5.373 0 12c0 5.302 3.438 9.8 8.207 11.387.6.113.82-.26.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61-.546-1.386-1.332-1.755-1.332-1.755-1.087-.744.084-.729.084-.729 1.205.085 1.84 1.236 1.84 1.236 1.07 1.835 2.809 1.305 3.493.997.108-.776.42-1.305.763-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 21.795 24 17.295 24 12c0-6.627-5.373-12-12-12z" fill="#24292F" />
                </svg>
                <span className="hidden sm:inline whitespace-nowrap">GitHub</span>
              </button>

              <button type="button" onClick={() => handleOAuth("slack")} className="h-12 bg-white border-2 border-gray-300 shadow-sm transition-all text-gray-700 hover:bg-gray-50 hover:text-purple-600 rounded-lg flex items-center justify-center gap-2 border bg-background inline-flex whitespace-nowrap text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50">
                <Image src="/slack.png" alt="Slack" width={18} height={18} className="object-contain" />
                <span className="hidden sm:inline whitespace-nowrap">Slack</span>
              </button>
            </div>
          </form>

          <div className="p-0 mt-6">
            <p className="text-sm text-center text-gray-500 w-full">
              Don&apos;t have an account?{" "}
              <Link href="/register" className="text-blue-600 hover:underline font-medium">
                Sign up
              </Link>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
