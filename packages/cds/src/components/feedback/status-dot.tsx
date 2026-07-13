import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const statusDotVariants = cva("inline-block h-2 w-2 rounded-full", {
  variants: {
    status: {
      online: "bg-success",
      offline: "bg-muted-foreground",
      busy: "bg-warning",
      error: "bg-destructive",
      pending: "bg-info animate-cds-pulse-soft",
    },
  },
  defaultVariants: { status: "online" },
});

export interface StatusDotProps
  extends React.HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof statusDotVariants> {
  label?: string;
}

export function StatusDot({ status, className, label, ...props }: StatusDotProps) {
  return (
    <span
      className={cn(statusDotVariants({ status }), className)}
      role="status"
      aria-label={label ?? status ?? "status"}
      {...props}
    />
  );
}
