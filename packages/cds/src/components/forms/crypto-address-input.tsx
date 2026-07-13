"use client";

import * as React from "react";
import { Copy, Check } from "lucide-react";
import { Input, type InputProps } from "@/components/ui/input";
import { cn } from "@/lib/utils";

export interface CryptoAddressInputProps extends Omit<InputProps, "onCopy"> {
  onAddressCopy?: (value: string) => void;
}

export const CryptoAddressInput = React.forwardRef<HTMLInputElement, CryptoAddressInputProps>(
  ({ className, value, onAddressCopy, ...props }, ref) => {
    const [copied, setCopied] = React.useState(false);
    const str = String(value ?? "");

    const handleCopy = async () => {
      if (!str) return;
      await navigator.clipboard.writeText(str);
      setCopied(true);
      onAddressCopy?.(str);
      setTimeout(() => setCopied(false), 1500);
    };

    return (
      <div className="relative">
        <Input
          ref={ref}
          value={value}
          className={cn("pr-10 font-mono text-xs sm:text-sm", className)}
          spellCheck={false}
          autoComplete="off"
          {...props}
        />
        <button
          type="button"
          onClick={handleCopy}
          className="absolute right-2 top-1/2 -translate-y-1/2 rounded-md p-1 text-muted-foreground hover:text-foreground"
          aria-label="Copy address"
        >
          {copied ? <Check className="h-4 w-4 text-success" /> : <Copy className="h-4 w-4" />}
        </button>
      </div>
    );
  },
);
CryptoAddressInput.displayName = "CryptoAddressInput";
