export function formatDate(value?: string | null, opts?: Intl.DateTimeFormatOptions): string {
  if (!value) return "—";
  try {
    return new Date(value).toLocaleString(undefined, opts ?? {
      dateStyle: "medium",
      timeStyle: "short",
    });
  } catch {
    return value;
  }
}

export function formatRelative(value?: string | null): string {
  if (!value) return "—";
  try {
    const date = new Date(value);
    const diff = Date.now() - date.getTime();
    const mins = Math.floor(diff / 60_000);
    if (mins < 1) return "Just now";
    if (mins < 60) return `${mins}m ago`;
    const hours = Math.floor(mins / 60);
    if (hours < 24) return `${hours}h ago`;
    const days = Math.floor(hours / 24);
    if (days < 30) return `${days}d ago`;
    return formatDate(value, { dateStyle: "medium" });
  } catch {
    return value;
  }
}

export function displayName(user?: {
  display_name?: string | null;
  username?: string | null;
  email?: string;
} | null): string {
  if (!user) return "User";
  return user.display_name || user.username || user.email?.split("@")[0] || "User";
}

export function initials(user?: {
  display_name?: string | null;
  username?: string | null;
  email?: string;
} | null): string {
  const name = displayName(user);
  return name.slice(0, 2).toUpperCase();
}

export function humanizeAction(action: string): string {
  return action
    .replace(/[._]/g, " ")
    .replace(/\b\w/g, (c) => c.toUpperCase());
}

export function profileCompletion(user?: {
  display_name?: string | null;
  username?: string | null;
  avatar_url?: string | null;
  country?: string | null;
  is_verified?: boolean;
} | null): { percent: number; missing: string[] } {
  if (!user) return { percent: 0, missing: ["Profile"] };
  const checks: { key: string; ok: boolean }[] = [
    { key: "Display name", ok: !!user.display_name },
    { key: "Username", ok: !!user.username },
    { key: "Avatar", ok: !!user.avatar_url },
    { key: "Country", ok: !!user.country },
    { key: "Email verified", ok: !!user.is_verified },
  ];
  const done = checks.filter((c) => c.ok).length;
  return {
    percent: Math.round((done / checks.length) * 100),
    missing: checks.filter((c) => !c.ok).map((c) => c.key),
  };
}
