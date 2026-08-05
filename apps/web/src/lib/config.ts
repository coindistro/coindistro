export const appConfig = {
  name: "Coindistro",
  apiBaseUrl:
    process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") ||
    "http://localhost:8080",
  siteUrl: process.env.NEXT_PUBLIC_SITE_URL || "http://localhost:3000",
  accessTokenKey: "coindistro_access_token",
  refreshTokenKey: "coindistro_refresh_token",
  userKey: "coindistro_user",
  /**
   * Optional Paystack public key for client-side popup/inline flows.
   * Transaction initialization always uses the backend secret key.
   * Set NEXT_PUBLIC_PAYSTACK_PUBLIC_KEY from PAYSTACK_PUBLIC_KEY at deploy time.
   */
  paystackPublicKey: process.env.NEXT_PUBLIC_PAYSTACK_PUBLIC_KEY || "",
} as const;
