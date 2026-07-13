import * as React from "react";
import { cn } from "@/lib/utils";
import { cdsTypography, type CdsTypographyVariant } from "@/tokens";

export interface TypographyProps extends React.HTMLAttributes<HTMLElement> {
  variant?: CdsTypographyVariant;
  as?: keyof JSX.IntrinsicElements;
}

export function Typography({
  variant = "body",
  as,
  className,
  children,
  ...props
}: TypographyProps) {
  const Comp = (as ??
    (variant === "display" || variant.startsWith("h")
      ? variant === "display"
        ? "h1"
        : variant
      : "p")) as keyof JSX.IntrinsicElements;

  return React.createElement(
    Comp,
    {
      className: cn(cdsTypography[variant], className),
      ...props,
    },
    children,
  );
}
