"use client";

import * as React from "react";
import { cn } from "@coindistro/cds";

/** Lightweight SVG sparkline — no chart library, zero recharts cost. */
export const MiniSparkline = React.memo(function MiniSparkline({
  data,
  positive = true,
  className,
  width = 72,
  height = 28,
}: {
  data: number[];
  positive?: boolean;
  className?: string;
  width?: number;
  height?: number;
}) {
  const path = React.useMemo(() => {
    if (!data.length) return "";
    const min = Math.min(...data);
    const max = Math.max(...data);
    const range = max - min || 1;
    const step = width / Math.max(data.length - 1, 1);
    return data
      .map((value, index) => {
        const x = index * step;
        const y = height - ((value - min) / range) * (height - 4) - 2;
        return `${index === 0 ? "M" : "L"}${x.toFixed(1)} ${y.toFixed(1)}`;
      })
      .join(" ");
  }, [data, height, width]);

  return (
    <svg
      width={width}
      height={height}
      viewBox={`0 0 ${width} ${height}`}
      className={cn("overflow-visible", className)}
      aria-hidden
    >
      <path
        d={path}
        fill="none"
        stroke={positive ? "rgb(16 185 129)" : "rgb(244 63 94)"}
        strokeWidth="1.75"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
});
