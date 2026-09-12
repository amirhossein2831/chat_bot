import { api } from "../../lib/apiClient";

type FactWire = {
  id: string;
  text: string;
  pinned: boolean;
  created_at: string;
};

export type Fact = {
  id: string;
  text: string;
  pinned: boolean;
  createdAt: string;
};

const toFact = (wire: FactWire): Fact => ({
  id: wire.id,
  text: wire.text,
  pinned: wire.pinned,
  createdAt: wire.created_at,
});

export async function fetchFacts(): Promise<Fact[]> {
  const wire = await api.get<FactWire[]>("/memory");
  return (wire ?? []).map(toFact);
}

export async function addFact(text: string): Promise<Fact> {
  return toFact(await api.post<FactWire>("/memory", { text }));
}

export const deleteFact = (id: string): Promise<void> => api.del<void>(`/memory/${id}`);

export async function fetchInstructions(): Promise<string> {
  const wire = await api.get<{ instructions: string }>("/settings");
  return wire.instructions ?? "";
}

export const saveInstructions = (instructions: string): Promise<void> =>
  api.put<void>("/settings", { instructions });
