export const appConfig = {
  name: "Coindistro",
  apiBaseUrl:
    process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") ||
    "http://localhost:8080",
  siteUrl: process.env.NEXT_PUBLIC_SITE_URL || "http://localhost:3000",
  accessTokenKey: "coindistro_access_token",
  refreshTokenKey: "coindistro_refresh_token",
  userKey: "coindistro_user",
} as const;
