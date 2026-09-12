import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { addFact, deleteFact, fetchFacts, fetchInstructions, saveInstructions } from "./memoryApi";

const factsKey = ["memory", "facts"] as const;
const settingsKey = ["memory", "settings"] as const;

export function useFacts() {
  return useQuery({ queryKey: factsKey, queryFn: fetchFacts });
}

export function useInstructions() {
  return useQuery({ queryKey: settingsKey, queryFn: fetchInstructions });
}

export function useSaveInstructions() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: saveInstructions,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: settingsKey }),
  });
}

export function useAddFact() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: addFact,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: factsKey }),
  });
}

export function useDeleteFact() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteFact,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: factsKey }),
  });
}
