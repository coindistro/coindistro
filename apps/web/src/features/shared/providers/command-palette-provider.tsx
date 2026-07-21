"use client";

import * as React from "react";
import { useRouter } from "next/navigation";
import { cn } from "@coindistro/cds";
import { Search } from "lucide-react";

interface CommandItemDef {
  label: string;
  href: string;
  icon: React.ReactNode;
  group: string;
}

const defaultItems: CommandItemDef[] = [
  { label: "Dashboard", href: "/app/dashboard", icon: <span className="text-primary">D</span>, group: "Navigation" },
  { label: "Markets", href: "/app/markets", icon: <span className="text-primary">M</span>, group: "Navigation" },
  { label: "Trade", href: "/app/trade", icon: <span className="text-primary">T</span>, group: "Navigation" },
  { label: "P2P", href: "/app/p2p", icon: <span className="text-primary">P</span>, group: "Navigation" },
  { label: "Earn", href: "/app/earn", icon: <span className="text-primary">E</span>, group: "Navigation" },
  { label: "Academy", href: "/app/academy", icon: <span className="text-primary">A</span>, group: "Navigation" },
  { label: "Signals", href: "/app/signals", icon: <span className="text-primary">S</span>, group: "Navigation" },
  { label: "AI Bots", href: "/app/ai-bots", icon: <span className="text-primary">B</span>, group: "Navigation" },
  { label: "Wallet", href: "/app/wallet", icon: <span className="text-primary">W</span>, group: "Navigation" },
  { label: "Merchant", href: "/app/merchant", icon: <span className="text-primary">M</span>, group: "Navigation" },
  { label: "Pay", href: "/app/pay", icon: <span className="text-primary">P</span>, group: "Navigation" },
  { label: "Referrals", href: "/app/referrals", icon: <span className="text-primary">R</span>, group: "Navigation" },
  { label: "Notifications", href: "/app/notifications", icon: <span className="text-primary">N</span>, group: "Navigation" },
  { label: "Profile", href: "/app/profile", icon: <span className="text-primary">U</span>, group: "Navigation" },
  { label: "Settings", href: "/app/settings", icon: <span className="text-primary">S</span>, group: "Navigation" },
];

interface CommandPaletteContextValue {
  open: boolean;
  setOpen: (open: boolean) => void;
}

const CommandPaletteContext = React.createContext<CommandPaletteContextValue | null>(null);

export function CommandPaletteProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [open, setOpen] = React.useState(false);
  const [query, setQuery] = React.useState("");
  const [selectedIndex, setSelectedIndex] = React.useState(0);
  const inputRef = React.useRef<HTMLInputElement>(null);

  // Keyboard shortcut: Cmd+K / Ctrl+K
  React.useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((prev) => !prev);
      }
    };
    document.addEventListener("keydown", down);
    return () => document.removeEventListener("keydown", down);
  }, []);

  // Focus input when opened
  React.useEffect(() => {
    if (open) {
      setTimeout(() => inputRef.current?.focus(), 50);
      setQuery("");
      setSelectedIndex(0);
    }
  }, [open]);

  const filtered = React.useMemo(() => {
    if (!query.trim()) return defaultItems;
    const q = query.toLowerCase();
    return defaultItems.filter(
      (item) =>
        item.label.toLowerCase().includes(q) ||
        item.group.toLowerCase().includes(q),
    );
  }, [query]);

  const grouped = React.useMemo(() => {
    const map = new Map<string, CommandItemDef[]>();
    for (const item of filtered) {
      const group = map.get(item.group) ?? [];
      group.push(item);
      map.set(item.group, group);
    }
    return Array.from(map.entries());
  }, [filtered]);

  const flatItems = React.useMemo(() => filtered, [filtered]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setSelectedIndex((i) => Math.min(i + 1, flatItems.length - 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setSelectedIndex((i) => Math.max(i - 1, 0));
    } else if (e.key === "Enter" && flatItems[selectedIndex]) {
      e.preventDefault();
      setOpen(false);
      router.push(flatItems[selectedIndex].href);
    } else if (e.key === "Escape") {
      setOpen(false);
    }
  };

  return (
    <CommandPaletteContext.Provider value={{ open, setOpen }}>
      {children}
      {open && (
        <div className="fixed inset-0 z-[700] flex items-start justify-center pt-[15vh]">
          {/* Backdrop */}
          <button
            type="button"
            className="absolute inset-0 bg-black/60 backdrop-blur-sm"
            onClick={() => setOpen(false)}
            aria-label="Close command palette"
          />
          {/* Dialog */}
          <div
            className="relative w-full max-w-lg rounded-xl border bg-card shadow-2xl"
            role="dialog"
            aria-modal="true"
            aria-label="Command palette"
          >
            {/* Search input */}
            <div className="flex items-center gap-3 border-b px-4 py-3">
              <Search className="h-4 w-4 shrink-0 text-muted-foreground" />
              <input
                ref={inputRef}
                type="text"
                className="flex-1 bg-transparent text-sm outline-none placeholder:text-muted-foreground"
                placeholder="Search modules, settings, or pages…"
                value={query}
                onChange={(e) => {
                  setQuery(e.target.value);
                  setSelectedIndex(0);
                }}
                onKeyDown={handleKeyDown}
                aria-label="Search"
              />
              <kbd className="hidden shrink-0 rounded border bg-muted px-1.5 text-[10px] font-medium text-muted-foreground sm:inline-block">
                ESC
              </kbd>
            </div>
            {/* Results */}
            <div className="max-h-80 overflow-y-auto p-2">
              {grouped.length === 0 && (
                <p className="py-6 text-center text-sm text-muted-foreground">
                  No results found.
                </p>
              )}
              {grouped.map(([group, items]) => (
                <div key={group} className="mb-2">
                  <p className="px-2 py-1 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                    {group}
                  </p>
                  {items.map((item) => {
                    const globalIdx = flatItems.indexOf(item);
                    return (
                      <button
                        key={item.href}
                        type="button"
                        className={cn(
                          "flex w-full items-center gap-3 rounded-lg px-3 py-2 text-left text-sm transition-colors",
                          globalIdx === selectedIndex
                            ? "bg-accent text-accent-foreground"
                            : "text-foreground hover:bg-muted",
                        )}
                        onClick={() => {
                          setOpen(false);
                          router.push(item.href);
                        }}
                        onMouseEnter={() => setSelectedIndex(globalIdx)}
                      >
                        <span className="flex h-6 w-6 items-center justify-center rounded-md bg-muted text-xs font-bold">
                          {item.icon}
                        </span>
                        <span className="flex-1">{item.label}</span>
                      </button>
                    );
                  })}
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </CommandPaletteContext.Provider>
  );
}

export function useCommandPalette() {
  const ctx = React.useContext(CommandPaletteContext);
  if (!ctx) throw new Error("useCommandPalette must be used within CommandPaletteProvider");
  return ctx;
}