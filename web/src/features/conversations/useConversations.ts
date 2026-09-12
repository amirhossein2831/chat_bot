import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import {
  archiveConversation,
  createConversation,
  listConversations,
  renameConversation,
} from "./conversationsApi";

export const conversationsKey = (query: string) => ["conversations", query] as const;

export function useConversations(query: string) {
  return useQuery({
    queryKey: conversationsKey(query),
    queryFn: () => listConversations(query),
    // Search results should feel immediate when a query is re-typed.
    placeholderData: (previous) => previous,
  });
}

function useInvalidateConversations() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: ["conversations"] });
}

export function useCreateConversation() {
  const invalidate = useInvalidateConversations();

  return useMutation({
    mutationFn: ({ title, model }: { title: string; model: string }) =>
      createConversation(title, model),
    onSuccess: invalidate,
  });
}

export function useRenameConversation() {
  const invalidate = useInvalidateConversations();

  return useMutation({
    mutationFn: ({ id, title }: { id: string; title: string }) => renameConversation(id, title),
    onSuccess: invalidate,
  });
}

export function useArchiveConversation() {
  const invalidate = useInvalidateConversations();

  return useMutation({
    mutationFn: (id: string) => archiveConversation(id),
    onSuccess: invalidate,
  });
}
