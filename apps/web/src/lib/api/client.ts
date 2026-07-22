import { appConfig } from "@/lib/config";
import { ApiError, type ApiResponse } from "@/lib/api/types";

type HttpMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE";

export interface RequestOptions {
  method?: HttpMethod;
  body?: unknown;
  headers?: Record<string, string>;
  auth?: boolean;
  signal?: AbortSignal;
  /** Skip automatic refresh retry once. */
  _retry?: boolean;
}

let accessToken: string | null = null;
let refreshToken: string | null = null;
let refreshPromise: Promise<boolean> | null = null;
let onUnauthorized: (() => void) | null = null;

export function setAuthTokens(access: string | null, refresh: string | null) {
  accessToken = access;
  refreshToken = refresh;
  if (typeof window !== "undefined") {
    if (access) localStorage.setItem(appConfig.accessTokenKey, access);
    else localStorage.removeItem(appConfig.accessTokenKey);
    if (refresh) localStorage.setItem(appConfig.refreshTokenKey, refresh);
    else localStorage.removeItem(appConfig.refreshTokenKey);
  }
}

export function loadAuthTokensFromStorage() {
  if (typeof window === "undefined") return;
  accessToken = localStorage.getItem(appConfig.accessTokenKey);
  refreshToken = localStorage.getItem(appConfig.refreshTokenKey);
}

export function getAccessToken() {
  return accessToken;
}

export function setUnauthorizedHandler(handler: () => void) {
  onUnauthorized = handler;
}

async function tryRefresh(): Promise<boolean> {
  if (!refreshToken) return false;
  if (refreshPromise) return refreshPromise;

  refreshPromise = (async () => {
    try {
      const res = await fetch(`${appConfig.apiBaseUrl}/api/v1/auth/refresh`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });
      if (!res.ok) return false;
      const json = (await res.json()) as ApiResponse<{
        access_token: string;
        refresh_token: string;
      }>;
      if (!json.success || !json.data?.access_token) return false;
      setAuthTokens(
        json.data.access_token,
        json.data.refresh_token || refreshToken,
      );
      return true;
    } catch {
      return false;
    } finally {
      refreshPromise = null;
    }
  })();

  return refreshPromise;
}

export async function apiRequest<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const {
    method = "GET",
    body,
    headers = {},
    auth = true,
    signal,
    _retry = false,
  } = options;

  const url = path.startsWith("http")
    ? path
    : `${appConfig.apiBaseUrl}${path.startsWith("/") ? path : `/${path}`}`;

  const reqHeaders: Record<string, string> = {
    Accept: "application/json",
    ...headers,
  };
  if (body !== undefined) {
    reqHeaders["Content-Type"] = "application/json";
  }
  if (auth && accessToken) {
    reqHeaders.Authorization = `Bearer ${accessToken}`;
  }

  let res: Response;
  try {
    res = await fetch(url, {
      method,
      headers: reqHeaders,
      body: body !== undefined ? JSON.stringify(body) : undefined,
      signal,
    });
  } catch (err) {
    throw new ApiError(
      0,
      "NETWORK_ERROR",
      err instanceof Error ? err.message : "Network request failed",
    );
  }

  // Token expired — attempt refresh once
  if (res.status === 401 && auth && !_retry) {
    const ok = await tryRefresh();
    if (ok) {
      return apiRequest<T>(path, { ...options, _retry: true });
    }
    onUnauthorized?.();
    throw new ApiError(401, "UNAUTHORIZED", "Session expired. Please sign in again.");
  }

  let json: ApiResponse<T> | null = null;
  const text = await res.text();
  if (text) {
    try {
      json = JSON.parse(text) as ApiResponse<T>;
    } catch {
      throw new ApiError(res.status, "INVALID_JSON", "Invalid server response");
    }
  }

  if (!res.ok) {
    throw new ApiError(
      res.status,
      json?.error?.code || `HTTP_${res.status}`,
      json?.error?.message || json?.message || res.statusText || "Request failed",
      json?.error?.details,
    );
  }

  if (json && json.success === false) {
    throw new ApiError(
      res.status,
      json.error?.code || "API_ERROR",
      json.error?.message || json.message || "Request failed",
      json.error?.details,
    );
  }

  return (json?.data !== undefined ? json.data : (json as unknown as T)) as T;
}

export const api = {
  get: <T>(path: string, opts?: Omit<RequestOptions, "method" | "body">) =>
    apiRequest<T>(path, { ...opts, method: "GET" }),
  post: <T>(path: string, body?: unknown, opts?: Omit<RequestOptions, "method" | "body">) =>
    apiRequest<T>(path, { ...opts, method: "POST", body }),
  put: <T>(path: string, body?: unknown, opts?: Omit<RequestOptions, "method" | "body">) =>
    apiRequest<T>(path, { ...opts, method: "PUT", body }),
  patch: <T>(path: string, body?: unknown, opts?: Omit<RequestOptions, "method" | "body">) =>
    apiRequest<T>(path, { ...opts, method: "PATCH", body }),
  delete: <T>(path: string, opts?: Omit<RequestOptions, "method" | "body">) =>
    apiRequest<T>(path, { ...opts, method: "DELETE" }),
};
