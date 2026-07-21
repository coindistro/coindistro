import { Badge } from "../../components/ui/badge";

export type KycStatus = "none" | "pending" | "approved" | "rejected" | "needs_review";

const map: Record<
  KycStatus,
  { label: string; variant: "muted" | "warning" | "success" | "danger" | "info" }
> = {
  none: { label: "Not started", variant: "muted" },
  pending: { label: "Pending", variant: "warning" },
  approved: { label: "Verified", variant: "success" },
  rejected: { label: "Rejected", variant: "danger" },
  needs_review: { label: "Needs review", variant: "info" },
};

export function KycStatusBadge({ status }: { status: KycStatus }) {
  const cfg = map[status];
  return <Badge variant={cfg.variant}>{cfg.label}</Badge>;
}
