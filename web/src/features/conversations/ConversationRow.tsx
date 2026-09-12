import { useState } from "react";
import { useTranslation } from "react-i18next";
import { NavLink } from "react-router-dom";

import type { Conversation } from "./conversationsApi";
import { useArchiveConversation, useRenameConversation } from "./useConversations";

type Props = {
  conversation: Conversation;
  onOpen: () => void;
};

export function ConversationRow({ conversation, onOpen }: Props) {
  const { t } = useTranslation();
  const [menuOpen, setMenuOpen] = useState(false);
  const [renaming, setRenaming] = useState(false);
  const [title, setTitle] = useState(conversation.title);

  const rename = useRenameConversation();
  const archive = useArchiveConversation();

  const commitRename = () => {
    setRenaming(false);
    const next = title.trim();
    if (next && next !== conversation.title) {
      rename.mutate({ id: conversation.id, title: next });
    } else {
      setTitle(conversation.title);
    }
  };

  if (renaming) {
    return (
      <input
        autoFocus
        value={title}
        onChange={(event) => setTitle(event.target.value)}
        onBlur={commitRename}
        onKeyDown={(event) => {
          if (event.key === "Enter") commitRename();
          if (event.key === "Escape") {
            setTitle(conversation.title);
            setRenaming(false);
          }
        }}
        className="h-10 w-full rounded-[12px] bg-raised px-3 text-[14px] outline-none ring-2 ring-accent"
      />
    );
  }

  return (
    <div className="group relative flex min-w-0 items-center">
      <NavLink
        to={`/c/${conversation.id}`}
        onClick={onOpen}
        className={({ isActive }) =>
          [
            // min-w-0 lets the title truncate: without it this item refuses to
            // shrink below its text and a long title scrolls the sidebar.
            "flex h-10 min-w-0 flex-1 items-center rounded-[12px] px-3 pe-9 transition-colors duration-150",
            isActive ? "bg-raised font-medium" : "hover:bg-raised/70",
          ].join(" ")
        }
      >
        {/* dir="auto" so an English title truncates at its own end: laid out
            RTL it would be clipped at the front and show only its tail. */}
        <span dir="auto" className="truncate text-[14px]">
          {conversation.title || t("nav.newChat")}
        </span>
      </NavLink>

      <button
        type="button"
        onClick={() => setMenuOpen((current) => !current)}
        aria-label={t("conversation.menu")}
        // Always reachable on touch, where there is no hover to reveal it.
        className="absolute end-1.5 flex size-7 items-center justify-center rounded-lg text-muted transition hover:bg-line/70 lg:opacity-0 lg:group-hover:opacity-100"
      >
        <svg width="15" height="15" viewBox="0 0 24 24" fill="currentColor">
          <circle cx="12" cy="5" r="1.6" />
          <circle cx="12" cy="12" r="1.6" />
          <circle cx="12" cy="19" r="1.6" />
        </svg>
      </button>

      {menuOpen && (
        <>
          <button
            type="button"
            aria-hidden
            tabIndex={-1}
            onClick={() => setMenuOpen(false)}
            className="fixed inset-0 z-10"
          />
          <div className="absolute top-full z-20 mt-1 flex w-40 flex-col rounded-[14px] bg-paper p-1.5 shadow-[var(--shadow-lift)] end-1.5">
            <button
              type="button"
              onClick={() => {
                setMenuOpen(false);
                setRenaming(true);
              }}
              className="rounded-[9px] px-3 py-2 text-start text-[13.5px] transition hover:bg-raised"
            >
              {t("conversation.rename")}
            </button>
            <button
              type="button"
              onClick={() => {
                setMenuOpen(false);
                archive.mutate(conversation.id);
              }}
              className="rounded-[9px] px-3 py-2 text-start text-[13.5px] text-red-600 transition hover:bg-raised dark:text-red-400"
            >
              {t("conversation.delete")}
            </button>
          </div>
        </>
      )}
    </div>
  );
}
