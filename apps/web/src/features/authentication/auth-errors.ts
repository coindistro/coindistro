import { ApiError } from "@/lib/api/types";

/**
 * Map backend auth failures to user-facing copy.
 * Prefer server message when present; fall back by status code.
 */
export function mapAuthError(
  error: unknown,
  fallback = "Authentication failed",
): string {
  if (error instanceof ApiError) {
    if (error.message && error.message.trim()) {
      // Prefer explicit backend messages for 401/403/423.
      if ([401, 403, 423].includes(error.status)) {
        return error.message;
      }
      if (error.status >= 500) {
        return "Unexpected server error";
      }
      return error.message;
    }

    switch (error.status) {
      case 401:
        return "Invalid email or password";
      case 403:
        return "Account not verified";
      case 423:
        return "Account locked";
      case 0:
        return "Network error. Check your connection and try again.";
      default:
        if (error.status >= 500) return "Unexpected server error";
        return fallback;
    }
  }

  if (error instanceof Error && error.message) {
    return error.message;
  }

  return fallback;
}

export function isUnverifiedAccountError(error: unknown): boolean {
  return (
    error instanceof ApiError &&
    (error.status === 403 ||
      error.code === "ACCOUNT_NOT_VERIFIED" ||
      /verif/i.test(error.message))
  );
}
