"use client";

import {
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  EmptyState,
  PageHeader,
  Skeleton,
  StatCard,
} from "@coindistro/cds";
import { Construction } from "lucide-react";

export interface PlaceholderPageProps {
  title: string;
  description: string;
  module: string;
  stats?: { title: string; value: string }[];
}

/** Production-quality module placeholder — ready for live API wiring. */
export function PlaceholderPage({
  title,
  description,
  module,
  stats = [
    { title: "Status", value: "Coming soon" },
    { title: "Module", value: module },
    { title: "API", value: "Ready" },
  ],
}: PlaceholderPageProps) {
  return (
    <div className="space-y-6 animate-cds-fade-in">
      <PageHeader
        title={title}
        description={description}
        actions={
          <Badge variant="info" className="capitalize">
            Phase 1 · Shell
          </Badge>
        }
      />

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {stats.map((s) => (
          <StatCard key={s.title} title={s.title} value={s.value} />
        ))}
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Module workspace</CardTitle>
          <CardDescription>
            This route is reserved for the {module} product. Backend teams can
            connect live endpoints without changing app structure.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-3 sm:grid-cols-3">
            <Skeleton className="h-24 w-full" />
            <Skeleton className="h-24 w-full" />
            <Skeleton className="h-24 w-full" />
          </div>
          <EmptyState
            icon={<Construction className="h-8 w-8" />}
            title={`${title} is under construction`}
            description="UI shell and navigation are ready. Feature logic ships in a later phase."
            actionLabel="View documentation"
            onAction={() => {
              window.open("/docs", "_self");
            }}
          />
          <div className="flex flex-wrap gap-2">
            <Button variant="outline" size="sm" disabled>
              Primary action
            </Button>
            <Button variant="ghost" size="sm" disabled>
              Secondary
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
