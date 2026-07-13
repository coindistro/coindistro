"use client";

import {
  Area,
  AreaChart as ReAreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { cn } from "@/lib/utils";

export interface ChartPoint {
  label: string;
  value: number;
}

export interface CdsAreaChartProps {
  data: ChartPoint[];
  className?: string;
  height?: number;
  color?: string;
  showGrid?: boolean;
}

export function CdsAreaChart({
  data,
  className,
  height = 240,
  color = "hsl(var(--cds-primary))",
  showGrid = true,
}: CdsAreaChartProps) {
  return (
    <div className={cn("w-full", className)} style={{ height }}>
      <ResponsiveContainer width="100%" height="100%">
        <ReAreaChart data={data} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
          {showGrid ? (
            <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
          ) : null}
          <XAxis dataKey="label" tickLine={false} axisLine={false} className="text-xs" />
          <YAxis tickLine={false} axisLine={false} className="text-xs" width={40} />
          <Tooltip
            contentStyle={{
              borderRadius: 8,
              border: "1px solid hsl(var(--cds-border))",
              background: "hsl(var(--cds-popover))",
            }}
          />
          <Area
            type="monotone"
            dataKey="value"
            stroke={color}
            fill={color}
            fillOpacity={0.15}
            strokeWidth={2}
          />
        </ReAreaChart>
      </ResponsiveContainer>
    </div>
  );
}
