"use client";

import {
  Badge,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  EmptyState,
  PageHeader,
  StatCard,
} from "@coindistro/cds";
import { Construction, Clock, ListChecks } from "lucide-react";

export interface ComingSoonProps {
  title: string;
  description: string;
  module: string;
  status?: string;
  expectedFeatures?: string[];
}

/** Professional coming-soon screen for modules not yet implemented. */
export function ComingSoon({
  title,
  description,
  module,
  status = "Planned",
  expectedFeatures = [
    "Live data from Coindistro services",
    "Role-aware actions and permissions",
    "Full CDS design system polish",
  ],
}: ComingSoonProps) {
  return (
    <div className="space-y-6 animate-cds-fade-in">
      <PageHeader
        title={title}
        description={description}
        actions={
          <Badge variant="info" className="capitalize">
            Coming soon
          </Badge>
        }
      />

      <div className="grid gap-4 sm:grid-cols-3">
        <StatCard
          title="Current status"
          value={status}
          description="Module lifecycle"
          icon={<Clock className="h-4 w-4" />}
        />
        <StatCard
          title="Module"
          value={module}
          description="Internal key"
          icon={<Construction className="h-4 w-4" />}
        />
        <StatCard
          title="Shell"
          value="Ready"
          description="Routing & navigation live"
          icon={<ListChecks className="h-4 w-4" />}
        />
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Expected features</CardTitle>
            <CardDescription>
              What this module will deliver when connected to backend services.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <ul className="space-y-2 text-sm text-muted-foreground">
              {expectedFeatures.map((f) => (
                <li key={f} className="flex items-start gap-2">
                  <span className="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full bg-primary" />
                  <span>{f}</span>
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Workspace</CardTitle>
            <CardDescription>
              Navigation and layout are production-ready. Business logic ships in a later milestone.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <EmptyState
              icon={<Construction className="h-8 w-8" />}
              title={`${title} is under construction`}
              description="This page is reserved for a future Coindistro product module. Authentication, RBAC, and the design system already apply."
            />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
