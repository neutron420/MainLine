"use client";

import { createContext, useCallback, useContext, useMemo } from "react";

import { useQueryClient } from "@tanstack/react-query";

import { useCurrentUser, useLogin, useLogout, useRegister } from "@/lib/api/hooks/use-auth";
import { authStore } from "@/lib/api/auth-store";
import type { User } from "@/lib/gen/common/v1/common_pb";

type AuthContextValue = {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<User | null>;
  register: (input: {
    email: string;
    password: string;
    displayName: string;
  }) => Promise<User | null>;
  logout: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const queryClient = useQueryClient();
  const { data: user, isLoading } = useCurrentUser();
  const loginMutation = useLogin();
  const registerMutation = useRegister();
  const logoutMutation = useLogout();

  const login = useCallback(
    async (email: string, password: string) => {
      const loggedInUser = await loginMutation.mutateAsync({ email, password });
      queryClient.invalidateQueries();
      return loggedInUser;
    },
    [loginMutation, queryClient],
  );

  const register = useCallback(
    async (input: { email: string; password: string; displayName: string }) => {
      const registeredUser = await registerMutation.mutateAsync(input);
      queryClient.invalidateQueries();
      return registeredUser;
    },
    [registerMutation, queryClient],
  );

  const logout = useCallback(async () => {
    await logoutMutation.mutateAsync();
    queryClient.clear();
  }, [logoutMutation, queryClient]);

  const value = useMemo<AuthContextValue>(
    () => ({
      user: user ?? null,
      isLoading,
      isAuthenticated: authStore.isAuthenticated() && user !== undefined,
      login,
      register,
      logout,
    }),
    [user, isLoading, login, register, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return ctx;
}
