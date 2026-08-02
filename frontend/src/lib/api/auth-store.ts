const ACCESS_TOKEN_KEY = "schemahub.access_token";
const REFRESH_TOKEN_KEY = "schemahub.refresh_token";

function isBrowser(): boolean {
  return typeof window !== "undefined";
}

export const authStore = {
  getAccessToken(): string | null {
    if (!isBrowser()) return null;
    return window.localStorage.getItem(ACCESS_TOKEN_KEY);
  },

  getRefreshToken(): string | null {
    if (!isBrowser()) return null;
    return window.localStorage.getItem(REFRESH_TOKEN_KEY);
  },

  setTokens(accessToken: string, refreshToken: string): void {
    if (!isBrowser()) return;
    window.localStorage.setItem(ACCESS_TOKEN_KEY, accessToken);
    window.localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
  },

  clear(): void {
    if (!isBrowser()) return;
    window.localStorage.removeItem(ACCESS_TOKEN_KEY);
    window.localStorage.removeItem(REFRESH_TOKEN_KEY);
  },

  isAuthenticated(): boolean {
    return this.getAccessToken() !== null;
  },
};
