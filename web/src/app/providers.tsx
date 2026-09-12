import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useEffect, type ReactNode } from "react";

import { applyDirection } from "../lib/direction";
import { useUiStore } from "./store";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: (count, error) => count < 2 && !String(error).includes("unauthorized"),
      staleTime: 30_000,
      // Mobile connections drop; refetching on focus is how a returning tab
      // heals without the user reloading.
      refetchOnWindowFocus: true,
    },
  },
});

export function Providers({ children }: { children: ReactNode }) {
  const locale = useUiStore((state) => state.locale);

  useEffect(() => applyDirection(locale), [locale]);

  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}
