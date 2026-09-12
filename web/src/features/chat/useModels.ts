import { useQuery } from "@tanstack/react-query";

import { api } from "../../lib/apiClient";

type ModelWire = {
  id: string;
  provider: string;
  tier: string;
  context_window: number;
  available: boolean;
};

export type Model = {
  id: string;
  provider: string;
  tier: "local" | "frontier";
  contextWindow: number;
  available: boolean;
};

async function fetchModels(): Promise<Model[]> {
  const wire = await api.get<ModelWire[]>("/models");
  return (wire ?? []).map((m) => ({
    id: m.id,
    provider: m.provider,
    tier: m.tier as Model["tier"],
    contextWindow: m.context_window,
    available: m.available,
  }));
}

export function useModels() {
  return useQuery({
    queryKey: ["models"],
    queryFn: fetchModels,
    // Availability changes when a provider goes down, so this cannot be cached
    // for long or the picker lies about what can be used.
    staleTime: 30_000,
    refetchInterval: 60_000,
  });
}

// The catalogue lives in the gateway config. A hardcoded id here would be a
// second copy of it, wrong the first time a model is renamed.
export function useDefaultModel(): string {
  const { data } = useModels();
  return data?.find((model) => model.available)?.id ?? data?.[0]?.id ?? "";
}
