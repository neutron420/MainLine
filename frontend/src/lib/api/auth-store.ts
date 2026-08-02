const ACCESS_TOKEN_KEY = "schemahub.access_token";
const REFRESH_TOKEN_KEY = "schemahub.refresh_token";

function isBrowser(): boolean {
  return typeof window !== "undefined";
}

function storage(remember: boolean): Storage {
  return remember ? window.localStorage : window.sessionStorage;
}

export const authStore = {
  getAccessToken(): string | null {
    if (!isBrowser()) return null;
    return (
      window.localStorage.getItem(ACCESS_TOKEN_KEY) ??
      window.sessionStorage.getItem(ACCESS_TOKEN_KEY)
    );
  },

  getRefreshToken(): string | null {
    if (!isBrowser()) return null;
    return (
      window.localStorage.getItem(REFRESH_TOKEN_KEY) ??
      window.sessionStorage.getItem(REFRESH_TOKEN_KEY)
    );
  },

  setTokens(accessToken: string, refreshToken: string, remember = true): void {
    if (!isBrowser()) return;
    storage(remember).setItem(ACCESS_TOKEN_KEY, accessToken);
    storage(remember).setItem(REFRESH_TOKEN_KEY, refreshToken);
  },

  clear(): void {
    if (!isBrowser()) return;
    window.localStorage.removeItem(ACCESS_TOKEN_KEY);
    window.localStorage.removeItem(REFRESH_TOKEN_KEY);
    window.sessionStorage.removeItem(ACCESS_TOKEN_KEY);
    window.sessionStorage.removeItem(REFRESH_TOKEN_KEY);
  },

  isAuthenticated(): boolean {
    return this.getAccessToken() !== null;
  },
};
