import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { fetchThread, switchBranch } from "./chatApi";

export const threadKey = (conversationId: string) => ["thread", conversationId] as const;

export function useThread(conversationId: string) {
  return useQuery({
    queryKey: threadKey(conversationId),
    queryFn: () => fetchThread(conversationId),
    enabled: Boolean(conversationId),
  });
}

export function useSwitchBranch(conversationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (messageId: string) => switchBranch(conversationId, messageId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: threadKey(conversationId) }),
  });
}
