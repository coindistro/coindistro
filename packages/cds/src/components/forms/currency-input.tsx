"use client";

import * as React from "react";
import { Input, type InputProps } from "@/components/ui/input";
import { cn } from "@/lib/utils";

export interface CurrencyInputProps extends Omit<InputProps, "type" | "onChange"> {
  currency?: string;
  onValueChange?: (value: number | undefined) => void;
}

export const CurrencyInput = React.forwardRef<HTMLInputElement, CurrencyInputProps>(
  ({ className, currency = "USD", onValueChange, ...props }, ref) => {
    return (
      <div className="relative">
        <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">
          {currency}
        </span>
        <Input
          ref={ref}
          type="number"
          inputMode="decimal"
          className={cn("pl-14", className)}
          onChange={(e) => {
            const v = e.target.value;
            onValueChange?.(v === "" ? undefined : Number(v));
          }}
          {...props}
        />
      </div>
    );
  },
);
CurrencyInput.displayName = "CurrencyInput";
