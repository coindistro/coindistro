import type { LucideIcon, LucideProps } from "lucide-react";
import { cn } from "../../lib/utils";

export type IconSize = "xs" | "sm" | "md" | "lg" | "xl";

const sizeMap: Record<IconSize, string> = {
  xs: "h-3 w-3",
  sm: "h-4 w-4",
  md: "h-5 w-5",
  lg: "h-6 w-6",
  xl: "h-8 w-8",
};

export interface IconProps extends LucideProps {
  icon: LucideIcon;
  size?: IconSize;
}

/** Standardized Lucide icon wrapper. */
export function Icon({ icon: Lucide, size = "md", className, ...props }: IconProps) {
  return <Lucide className={cn(sizeMap[size], "shrink-0", className)} aria-hidden {...props} />;
}
