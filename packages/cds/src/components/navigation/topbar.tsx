"use client";

import * as React from "react";
import { Menu } from "lucide-react";
import { cn } from "../../lib/utils";
import { Button } from "../../components/ui/button";

export interface TopbarProps {
  title?: string;
  onMenuClick?: () => void;
  search?: React.ReactNode;
  actions?: React.ReactNode;
  className?: string;
}

export function Topbar({ title, onMenuClick, search, actions, className }: TopbarProps) {
  return (
    <div className={cn("flex h-14 items-center gap-3 px-4 sm:px-6", className)}>
      {onMenuClick ? (
        <Button
          variant="ghost"
          size="icon-sm"
          className="lg:hidden"
          onClick={onMenuClick}
          aria-label="Open menu"
        >
          <Menu className="h-5 w-5" />
        </Button>
      ) : null}
      {title ? <h2 className="text-sm font-semibold sm:text-base">{title}</h2> : null}
      <div className="ml-auto flex flex-1 items-center justify-end gap-2 sm:flex-initial">
        {search ? <div className="hidden max-w-xs flex-1 md:block">{search}</div> : null}
        {actions}
      </div>
    </div>
  );
}
