"use client";

import * as React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { CdsThemeProvider, TooltipProvider } from "@coindistro/cds";
import { AuthProvider } from "@/features/authentication/auth-provider";
import { ToastProvider } from "@/features/shared/providers/toast-provider";
import { CommandPaletteProvider } from "@/features/shared/providers/command-palette-provider";

function makeQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 30_000,
        retry: 1,
        refetchOnWindowFocus: false,
      },
    },
  });
}

export function AppProviders({ children }: { children: React.ReactNode }) {
  const [queryClient] = React.useState(() => makeQueryClient());

  return (
    <CdsThemeProvider defaultTheme="dark" enableSystem>
      <QueryClientProvider client={queryClient}>
        <TooltipProvider delayDuration={200}>
          <AuthProvider>
            <ToastProvider>
              <CommandPaletteProvider>{children}</CommandPaletteProvider>
            </ToastProvider>
          </AuthProvider>
        </TooltipProvider>
      </QueryClientProvider>
    </CdsThemeProvider>
  );
}
