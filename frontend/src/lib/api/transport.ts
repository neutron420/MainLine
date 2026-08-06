import { type Interceptor } from "@connectrpc/connect";
import { createGrpcWebTransport } from "@connectrpc/connect-web";

import { authStore } from "./auth-store";
import { isUnauthenticated } from "./errors";

export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

// ── Token refresh ──

let refreshPromise: Promise<boolean> | null = null;

/**
 * Exchanges the stored refresh token for a fresh pair. Concurrent 401s share
 * a single in-flight refresh so we never hammer the backend.
 */
export async function refreshTokens(): Promise<boolean> {
  const refreshToken = authStore.getRefreshToken();
  if (!refreshToken) return false;

  refreshPromise ??= doRefresh(refreshToken).finally(() => {
    refreshPromise = null;
  });
  return refreshPromise;
}

async function doRefresh(refreshToken: string): Promise<boolean> {
  const refreshUrl = `${API_BASE_URL}/schemahub.auth.v1.AuthService/RefreshToken`;

  try {
    const res = await fetch(refreshUrl, {
      method: "POST",
      headers: {
        "Content-Type": "application/grpc-web+json",
        "X-Grpc-Web": "1",
        "X-User-Agent": "grpc-web-javascript/0.1",
      },
      body: encodeGrpcWebFrame({ refreshToken }),
    });

    if (!res.ok) return false;

    const body = await res.arrayBuffer();
    const decoded = decodeGrpcWebFrame(body);
    if (!decoded) return false;

    const parsed = JSON.parse(decoded) as {
      accessToken?: string;
      refreshToken?: string;
      expiresIn?: number;
    };

    if (!parsed.accessToken || !parsed.refreshToken) return false;

    authStore.setTokens(parsed.accessToken, parsed.refreshToken);
    return true;
  } catch {
    return false;
  }
}

// ── gRPC-Web frame helpers (raw refresh call) ──

function encodeGrpcWebFrame(message: object): ArrayBuffer {
  const payload = new TextEncoder().encode(JSON.stringify(message));
  const frame = new Uint8Array(5 + payload.length);
  new DataView(frame.buffer).setUint32(1, payload.length, false);
  frame.set(payload, 5);
  return frame.buffer as ArrayBuffer;
}

function decodeGrpcWebFrame(body: ArrayBuffer): string | null {
  const bytes = new Uint8Array(body);
  let offset = 0;
  while (offset < bytes.length) {
    if (bytes[offset] !== 0) return null;
    const length = new DataView(
      bytes.buffer,
      bytes.byteOffset + offset + 1,
    ).getUint32(0, false);
    offset += 5;
    if (offset + length > bytes.length) return null;
    return new TextDecoder().decode(bytes.slice(offset, offset + length));
  }
  return null;
}

// ── Auth interceptor ──

/**
 * Injects the bearer token on every request. When a call throws
 * Unauthenticated, tries a single token refresh + retry before surfacing
 * the error to the caller.
 */
const authInterceptor: Interceptor = (next) => async (req) => {
  const token = authStore.getAccessToken();
  if (token) {
    req.header.set("Authorization", `Bearer ${token}`);
  }

  try {
    return await next(req);
  } catch (err) {
    if (isUnauthenticated(err) && authStore.getRefreshToken()) {
      if (await refreshTokens()) {
        const freshToken = authStore.getAccessToken();
        if (freshToken) {
          req.header.set("Authorization", `Bearer ${freshToken}`);
          return next(req);
        }
      }
      authStore.clear();
    }
    throw err;
  }
};

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const inFlightRequests = new Map<string, Promise<any>>();

/**
 * Deduplicates in-flight identical unary RPC queries to eliminate redundant
 * network roundtrips when multiple components mount concurrently.
 */
const deduplicationInterceptor: Interceptor = (next) => async (req) => {
  if (req.stream) {
    return next(req);
  }

  const key = `${req.service.typeName}/${req.method.name}:${JSON.stringify(req.message)}`;
  const existing = inFlightRequests.get(key);
  if (existing) {
    return existing;
  }

  const promise = next(req).finally(() => {
    inFlightRequests.delete(key);
  });

  inFlightRequests.set(key, promise);
  return promise;
};

export const transport = createGrpcWebTransport({
  baseUrl: API_BASE_URL,
  useBinaryFormat: true,
  defaultTimeoutMs: 60_000,
  interceptors: [deduplicationInterceptor, authInterceptor],
});
