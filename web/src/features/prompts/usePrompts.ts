import { useQuery } from "@tanstack/react-query";

import { fetchPrompts } from "./promptsApi";

export function usePrompts(locale: string, category: string) {
  return useQuery({
    queryKey: ["prompts", locale, category],
    queryFn: () => fetchPrompts(locale, category),
    staleTime: 10 * 60_000,
  });
}
