import { api } from "../../lib/apiClient";

type PromptWire = {
  id: string;
  locale: string;
  category: string;
  title: string;
  body: string;
};

export type Prompt = {
  id: string;
  locale: string;
  category: string;
  title: string;
  body: string;
};

export async function fetchPrompts(locale: string, category: string): Promise<Prompt[]> {
  const params = new URLSearchParams({ locale });
  if (category) params.set("category", category);

  const wire = await api.get<PromptWire[]>(`/prompts?${params.toString()}`);
  return wire ?? [];
}
