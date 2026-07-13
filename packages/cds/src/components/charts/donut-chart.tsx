"use client";

import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from "recharts";
import { cn } from "@/lib/utils";

export interface DonutSlice {
  name: string;
  value: number;
  color?: string;
}

const DEFAULT_COLORS = [
  "hsl(var(--cds-chart-1))",
  "hsl(var(--cds-chart-2))",
  "hsl(var(--cds-chart-3))",
  "hsl(var(--cds-chart-4))",
  "hsl(var(--cds-chart-5))",
];

export function CdsDonutChart({
  data,
  className,
  height = 220,
  innerRadius = 60,
}: {
  data: DonutSlice[];
  className?: string;
  height?: number;
  innerRadius?: number;
}) {
  return (
    <div className={cn("w-full", className)} style={{ height }}>
      <ResponsiveContainer width="100%" height="100%">
        <PieChart>
          <Pie
            data={data}
            dataKey="value"
            nameKey="name"
            innerRadius={innerRadius}
            outerRadius={innerRadius + 36}
            paddingAngle={2}
          >
            {data.map((entry, i) => (
              <Cell
                key={entry.name}
                fill={entry.color ?? DEFAULT_COLORS[i % DEFAULT_COLORS.length]}
              />
            ))}
          </Pie>
          <Tooltip
            contentStyle={{
              borderRadius: 8,
              border: "1px solid hsl(var(--cds-border))",
              background: "hsl(var(--cds-popover))",
            }}
          />
        </PieChart>
      </ResponsiveContainer>
    </div>
  );
}
