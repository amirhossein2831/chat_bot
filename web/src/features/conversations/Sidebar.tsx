import { useState } from "react";
import { useTranslation } from "react-i18next";
import { NavLink, useNavigate } from "react-router-dom";

import { useUiStore } from "../../app/store";
import { useLogout, useMe } from "../auth/useAuth";
import { ConversationRow } from "./ConversationRow";
import { useConversations } from "./useConversations";
import { useDebounced } from "./useDebounced";

export function Sidebar() {
  const { t } = useTranslation();
  const navigate = useNavigate();

  const setSidebarOpen = useUiStore((state) => state.setSidebarOpen);

  const [query, setQuery] = useState("");
  const debounced = useDebounced(query);

  const { data: conversations } = useConversations(debounced);
  const { data: me } = useMe();
  const logout = useLogout();

  const startConversation = () => {
    setSidebarOpen(false);
    navigate("/");
  };

  return (
    <div className="flex h-full flex-col">
      <div className="flex flex-col gap-3.5 px-[18px] pt-5 pb-3.5">
        <div className="flex items-center gap-2.5">
          <div className="flex size-[26px] items-center justify-center rounded-[7px] bg-accent">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.1" strokeLinecap="round" strokeLinejoin="round">
              <path d="M21 11.5a8.4 8.4 0 0 1-9 8.4 8.9 8.9 0 0 1-3.8-.9L3 21l1.9-5.1A8.4 8.4 0 0 1 12 3a8.4 8.4 0 0 1 9 8.5z" />
            </svg>
          </div>
          <span className="text-[15.5px] font-semibold tracking-tight">{t("app.name")}</span>
        </div>

        <button
          type="button"
          onClick={startConversation}
          className="flex h-11 w-full items-center justify-center gap-1.5 rounded-[9px] border border-accent/25 bg-surface text-sm font-medium text-accent md:h-10"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
            <path d="M12 5v14M5 12h14" />
          </svg>
          {t("nav.newChat")}
        </button>

        <NavLink
          to="/prompts"
          onClick={() => setSidebarOpen(false)}
          className="flex h-11 w-full items-center gap-1.5 rounded-[9px] px-2.5 text-sm text-muted hover:bg-surface md:h-9"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M4 5h16M4 12h16M4 19h10" />
          </svg>
          {t("nav.prompts")}
        </NavLink>

        <div className="relative flex items-center">
          <svg
            width="15"
            height="15"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            className="pointer-events-none absolute start-3 text-muted"
          >
            <circle cx="11" cy="11" r="7" />
            <path d="m20 20-3.2-3.2" />
          </svg>
          <input
            type="search"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder={t("search.placeholder")}
            className="h-11 w-full rounded-[9px] border border-line bg-surface ps-9 pe-3 text-[13.5px] outline-none focus:border-accent md:h-9"
          />
        </div>
      </div>

      <nav className="flex min-h-0 flex-1 flex-col gap-0.5 overflow-y-auto overflow-x-hidden px-2.5 pb-4">
        {debounced && (
          <span className="px-2 pt-2 pb-1 text-[11.5px] font-semibold text-muted">
            {t("search.results", { count: conversations?.length ?? 0 })}
          </span>
        )}

        {conversations?.length === 0 && (
          <span className="px-2 py-3 text-[13px] text-muted">{t("search.empty")}</span>
        )}

        {conversations?.map((conversation) => (
          <ConversationRow
            key={conversation.id}
            conversation={conversation}
            onOpen={() => setSidebarOpen(false)}
          />
        ))}

      </nav>

      <div className="mt-auto flex items-center gap-2.5 border-t border-line px-[18px] py-3.5">
        <div className="flex size-[30px] items-center justify-center rounded-full bg-accent-soft text-[12.5px] font-semibold text-accent">
          {me?.email.slice(0, 1).toUpperCase() ?? "?"}
        </div>
        <span className="min-w-0 flex-1 truncate text-[13px]">{me?.email}</span>
        <button
          type="button"
          onClick={() => {
            logout.mutate(undefined, { onSuccess: () => navigate("/login") });
          }}
          aria-label={t("auth.logout")}
          title={t("auth.logout")}
          className="flex size-11 items-center justify-center rounded-lg text-muted hover:bg-surface md:size-8"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M14 20H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h8" />
            <path d="M17 15l4-3-4-3" />
            <path d="M21 12H10" />
          </svg>
        </button>

        <NavLink
          to="/memory"
          onClick={() => setSidebarOpen(false)}
          aria-label={t("nav.memory")}
          className="flex size-11 items-center justify-center rounded-lg text-muted hover:bg-surface md:size-8"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <circle cx="12" cy="12" r="2.6" />
            <path d="M19.4 15a1.6 1.6 0 0 0 .3 1.8l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.6 1.6 0 0 0-2.7 1.1V21a2 2 0 1 1-4 0v-.1A1.6 1.6 0 0 0 8 19.4a1.6 1.6 0 0 0-1.8.3l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1a1.6 1.6 0 0 0-1.1-2.7H2a2 2 0 1 1 0-4h.1A1.6 1.6 0 0 0 3.6 8a1.6 1.6 0 0 0-.3-1.8l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1a1.6 1.6 0 0 0 1.8.3H8a1.6 1.6 0 0 0 1-1.5V2a2 2 0 1 1 4 0v.1a1.6 1.6 0 0 0 1 1.5 1.6 1.6 0 0 0 1.8-.3l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.6 1.6 0 0 0-.3 1.8V8a1.6 1.6 0 0 0 1.5 1H22a2 2 0 1 1 0 4h-.1a1.6 1.6 0 0 0-1.5 1z" />
          </svg>
        </NavLink>
      </div>
    </div>
  );
}
