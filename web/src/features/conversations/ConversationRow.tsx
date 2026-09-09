import { useState } from "react";
import { useTranslation } from "react-i18next";
import { NavLink } from "react-router-dom";

import { useUiStore } from "../../app/store";
import { formatRelative } from "../../lib/format";
import type { Conversation } from "./conversationsApi";
import { useArchiveConversation, useRenameConversation } from "./useConversations";

type Props = {
  conversation: Conversation;
  onOpen: () => void;
};

export function ConversationRow({ conversation, onOpen }: Props) {
  const { t } = useTranslation();
  const locale = useUiStore((state) => state.locale);

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
        className="h-11 w-full rounded-[9px] border border-accent bg-surface px-2.5 text-[13.5px] outline-none md:h-10"
      />
    );
  }

  return (
    <div className="group relative flex items-center">
      <NavLink
        to={`/c/${conversation.id}`}
        onClick={onOpen}
        className={({ isActive }) =>
          [
            "flex min-h-11 flex-1 flex-col justify-center gap-0.5 rounded-[9px] px-2.5 py-2 pe-9 md:min-h-0",
            isActive
              ? "bg-surface shadow-[inset_0_0_0_1px_var(--color-accent-soft)]"
              : "hover:bg-surface/60",
          ].join(" ")
        }
      >
        <span className="truncate text-[13.5px]">{conversation.title || t("nav.newChat")}</span>
        <span className="text-xs text-muted">
          {formatRelative(conversation.createdAt, locale)}
        </span>
      </NavLink>

      <button
        type="button"
        onClick={() => setMenuOpen((current) => !current)}
        aria-label={t("conversation.menu")}
        // Always reachable on touch, where there is no hover to reveal it.
        className="absolute end-1 flex size-8 items-center justify-center rounded-md text-muted opacity-100 hover:bg-line/60 lg:opacity-0 lg:group-hover:opacity-100"
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
          <div className="absolute top-full z-20 mt-1 flex w-40 flex-col rounded-lg border border-line bg-surface p-1 shadow-lg end-1">
            <button
              type="button"
              onClick={() => {
                setMenuOpen(false);
                setRenaming(true);
              }}
              className="rounded-md px-3 py-2 text-start text-[13px] hover:bg-raised"
            >
              {t("conversation.rename")}
            </button>
            <button
              type="button"
              onClick={() => {
                setMenuOpen(false);
                archive.mutate(conversation.id);
              }}
              className="rounded-md px-3 py-2 text-start text-[13px] text-red-600 hover:bg-raised dark:text-red-400"
            >
              {t("conversation.delete")}
            </button>
          </div>
        </>
      )}
    </div>
  );
}
