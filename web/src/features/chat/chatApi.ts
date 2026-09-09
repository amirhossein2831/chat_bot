import { api } from "../../lib/apiClient";

type MessageWire = {
  id: string;
  parent_id: string | null;
  role: string;
  content: string;
  status: string;
  model: string;
  tokens_in: number;
  tokens_out: number;
  error?: string | null;
  created_at: string;
};

export type Message = {
  id: string;
  parentId: string | null;
  role: "user" | "assistant" | "system";
  content: string;
  status: "streaming" | "complete" | "failed";
  model: string;
  tokensIn: number;
  tokensOut: number;
  error: string | null;
  createdAt: string;
};

const toMessage = (wire: MessageWire): Message => ({
  id: wire.id,
  parentId: wire.parent_id,
  role: wire.role as Message["role"],
  content: wire.content,
  status: wire.status as Message["status"],
  model: wire.model,
  tokensIn: wire.tokens_in,
  tokensOut: wire.tokens_out,
  error: wire.error ?? null,
  createdAt: wire.created_at,
});

export async function fetchThread(conversationId: string): Promise<Message[]> {
  const wire = await api.get<{ messages: MessageWire[] }>(
    `/conversations/${conversationId}/messages`,
  );
  return (wire.messages ?? []).map(toMessage);
}

export const stopGeneration = (messageId: string): Promise<void> =>
  api.post<void>(`/chat/${messageId}/stop`);

export const switchBranch = (conversationId: string, messageId: string): Promise<void> =>
  api.post<void>(`/conversations/${conversationId}/branch`, { message_id: messageId });
