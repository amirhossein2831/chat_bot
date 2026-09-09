import { api } from "../../lib/apiClient";

type ConversationWire = {
  id: string;
  title: string;
  default_model: string;
  head_message_id: string | null;
  archived: boolean;
  created_at: string;
};

export type Conversation = {
  id: string;
  title: string;
  defaultModel: string;
  headMessageId: string | null;
  archived: boolean;
  createdAt: string;
};

const toConversation = (wire: ConversationWire): Conversation => ({
  id: wire.id,
  title: wire.title,
  defaultModel: wire.default_model,
  headMessageId: wire.head_message_id,
  archived: wire.archived,
  createdAt: wire.created_at,
});

export async function listConversations(query: string): Promise<Conversation[]> {
  const suffix = query ? `?q=${encodeURIComponent(query)}` : "";
  const wire = await api.get<ConversationWire[]>(`/conversations${suffix}`);
  return (wire ?? []).map(toConversation);
}

export async function createConversation(title: string, model: string): Promise<Conversation> {
  return toConversation(await api.post<ConversationWire>("/conversations", { title, model }));
}

export const renameConversation = (id: string, title: string): Promise<void> =>
  api.patch<void>(`/conversations/${id}`, { title });

export const archiveConversation = (id: string): Promise<void> =>
  api.del<void>(`/conversations/${id}`);
