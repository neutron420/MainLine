"use client";

import { useEffect, useRef } from "react";
import { useSearchParams, useRouter } from "next/navigation";

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
      // TODO: show error state via URL params or context
      router.push("/login?error=invalid_oauth_response");
      return;
    }

    // TODO: call HandleOAuthCallback RPC
    // On success: store tokens, redirect to /dashboard
    // If needs_linking: redirect to /settings/connections
    // If is_new_user: redirect to /welcome

    router.push("/login");
  }, [searchParams, router]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-gray-50 to-gray-100 p-4">
      <div className="w-full max-w-md bg-white rounded-2xl shadow-xl p-8 text-center border border-gray-100">
        <div className="animate-spin w-10 h-10 border-4 border-blue-500 border-t-transparent rounded-full mx-auto mb-4" />
        <p className="text-gray-500">Completing sign in...</p>
      </div>
    </div>
  );
}
