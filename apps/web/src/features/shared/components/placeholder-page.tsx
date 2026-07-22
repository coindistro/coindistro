"use client";

import {
  ComingSoon,
  type ComingSoonProps,
} from "@/features/shared/components/coming-soon";

export type PlaceholderPageProps = ComingSoonProps & {
  stats?: { title: string; value: string }[];
};

/** @deprecated Prefer ComingSoon — kept for compatibility. */
export function PlaceholderPage({
  title,
  description,
  module,
  expectedFeatures,
  status,
}: PlaceholderPageProps) {
  return (
    <ComingSoon
      title={title}
      description={description}
      module={module}
      expectedFeatures={expectedFeatures}
      status={status ?? "Coming soon"}
    />
  );
}
