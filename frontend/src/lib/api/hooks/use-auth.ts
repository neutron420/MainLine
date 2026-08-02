"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { authClient } from "@/lib/api/clients";
import { authStore } from "@/lib/api/auth-store";
import { getApiErrorMessage } from "@/lib/api/errors";
import type { User } from "@/lib/gen/common/v1/common_pb";
import type { OAuthIdentity } from "@/lib/gen/auth/v1/auth_service_pb";

const USER_KEY = ["auth", "me"] as const;
const IDENTITIES_KEY = ["auth", "identities"] as const;

export function useCurrentUser() {
  return useQuery({
    queryKey: USER_KEY,
    queryFn: async () => {
      const res = await authClient.getCurrentUser({});
      return res.user ?? null;
    },
    enabled: authStore.isAuthenticated(),
    staleTime: 60_000,
    retry: false,
  });
}

export function useLogin() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ email, password }: { email: string; password: string }) => {
      const res = await authClient.login({ email, password });
      if (!res.accessToken || !res.refreshToken) {
        throw new Error("Login failed: server did not return tokens");
      }
      authStore.setTokens(res.accessToken, res.refreshToken);
      return res.user ?? null;
    },
    onSuccess: (user) => {
      queryClient.setQueryData(USER_KEY, user);
    },
  });
}

export function useRegister() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      email,
      password,
      displayName,
    }: {
      email: string;
      password: string;
      displayName: string;
    }) => {
      const res = await authClient.register({
        email,
        password,
        displayName,
      });
      if (!res.accessToken || !res.refreshToken) {
        throw new Error("Registration failed: server did not return tokens");
      }
      authStore.setTokens(res.accessToken, res.refreshToken);
      return res.user ?? null;
    },
    onSuccess: (user) => {
      queryClient.setQueryData(USER_KEY, user);
    },
  });
}

export function useLogout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      const refreshToken = authStore.getRefreshToken();
      if (refreshToken) {
        try {
          await authClient.logout({ refreshToken });
        } catch {
          // best-effort server-side invalidation
        }
      }
      authStore.clear();
      queryClient.clear();
    },
  });
}

export function useUpdateProfile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: { displayName?: string; avatarUrl?: string }) => {
      const res = await authClient.updateUser({
        displayName: input.displayName ?? "",
        avatarUrl: input.avatarUrl ?? "",
      });
      return res.user ?? null;
    },
    onSuccess: (user) => {
      queryClient.setQueryData(USER_KEY, user);
    },
  });
}

export function useChangePassword() {
  return useMutation({
    mutationFn: async ({
      currentPassword,
      newPassword,
    }: {
      currentPassword: string;
      newPassword: string;
    }) => {
      await authClient.changePassword({ currentPassword, newPassword });
    },
  });
}

export function useForgotPassword() {
  return useMutation({
    mutationFn: async (email: string) => {
      await authClient.forgotPassword({ email });
    },
  });
}

export function useResetPassword() {
  return useMutation({
    mutationFn: async (input: { token: string; password: string }) => {
      await authClient.resetPassword({ token: input.token, password: input.password });
    },
  });
}

export function useVerifyEmail() {
  return useMutation({
    mutationFn: async (token: string) => {
      await authClient.verifyEmail({ token });
    },
  });
}

export function useSendVerificationEmail() {
  return useMutation({
    mutationFn: async (email: string) => {
      await authClient.sendVerificationEmail({ email });
    },
  });
}

export function useListLinkedIdentities() {
  return useQuery({
    queryKey: IDENTITIES_KEY,
    queryFn: async () => {
      const res = await authClient.listLinkedIdentities({});
      return res.identities;
    },
    staleTime: 60_000,
  });
}

export function useGetOAuthUrl() {
  return useMutation({
    mutationFn: async (input: {
      provider: string;
      redirectTo?: string;
      linking?: boolean;
    }): Promise<string> => {
      const res = await authClient.getOAuthURL({
        provider: input.provider,
        redirectTo: input.redirectTo ?? "",
        linking: input.linking ?? false,
      });
      return res.authUrl;
    },
  });
}

export function useUnlinkOAuthIdentity() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (provider: string) => {
      await authClient.unlinkOAuthIdentity({ provider });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: IDENTITIES_KEY });
    },
  });
}

export type { User, OAuthIdentity };

export function formatApiError(err: unknown): string {
  return getApiErrorMessage(err);
}
