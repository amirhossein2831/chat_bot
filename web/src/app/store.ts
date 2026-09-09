import { create } from "zustand";

import type { Locale } from "../lib/direction";

type UiState = {
  locale: Locale;
  sidebarOpen: boolean;
  drafts: Record<string, string>;
  setLocale: (locale: Locale) => void;
  setSidebarOpen: (open: boolean) => void;
  setDraft: (conversationId: string, text: string) => void;
};

// Server state lives in TanStack Query. This holds only what the server never
// sees: language, sidebar visibility, unsent drafts. Theme is not here:
// it follows the OS through a media query, with nothing to store.
export const useUiStore = create<UiState>((set) => ({
  locale: "fa",
  sidebarOpen: false,
  drafts: {},

  setLocale: (locale) => set({ locale }),

  setSidebarOpen: (sidebarOpen) => set({ sidebarOpen }),

  setDraft: (conversationId, text) =>
    set((state) => ({ drafts: { ...state.drafts, [conversationId]: text } })),
}));
