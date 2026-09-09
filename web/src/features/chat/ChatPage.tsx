import { useCallback, useEffect, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useParams } from "react-router-dom";

import { useUiStore } from "../../app/store";
import type { Message } from "./chatApi";
import { Composer } from "./Composer";
import { MessageList } from "./MessageList";
import { useChatStream } from "./useChatStream";
import { useUpload } from "../files/useUpload";
import { useSwitchBranch, useThread } from "./useThread";

export function ChatPage() {
  const { t } = useTranslation();
  const { id = "" } = useParams();

  const drafts = useUiStore((state) => state.drafts);
  const setDraft = useUiStore((state) => state.setDraft);
  const setSidebarOpen = useUiStore((state) => state.setSidebarOpen);

  const { data: messages, refetch } = useThread(id);
  const switchBranch = useSwitchBranch(id);

  const onFinished = useCallback(() => {
    void refetch();
  }, [refetch]);

  const stream = useChatStream(onFinished);
  const uploads = useUpload();

  // A generation that was still running when the page was closed is picked up
  // where it left off, replayed from the beginning rather than the remainder.
  useEffect(() => {
    const last = messages?.[messages.length - 1];
    if (last?.role === "assistant" && last.status === "streaming" && !stream.streaming) {
      stream.resume(last.id);
    }
    // Only re-evaluated when the thread changes, never on every stream tick.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [messages]);

  useEffect(() => {
    stream.reset();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id]);

  const draft = drafts[id] ?? "";

  const send = () => {
    const content = draft.trim();
    if (!content) return;

    setDraft(id, "");
    uploads.clear();
    void stream.send(id, content);
  };

  // Siblings share a parent: that is what makes an alternative answer
  // reachable instead of lost.
  const siblingsOf = useMemo(() => {
    const byParent = new Map<string | null, string[]>();
    for (const message of messages ?? []) {
      const existing = byParent.get(message.parentId) ?? [];
      byParent.set(message.parentId, [...existing, message.id]);
    }
    return (message: Message) => byParent.get(message.parentId) ?? [];
  }, [messages]);

  return (
    <>
      <header className="flex h-14 shrink-0 items-center gap-3 border-b border-line px-4 md:px-7">
        <button
          type="button"
          onClick={() => setSidebarOpen(true)}
          aria-label={t("nav.newChat")}
          className="-ms-2 flex size-11 items-center justify-center rounded-lg text-muted lg:hidden"
        >
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
            <path d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>

        <span className="truncate text-[14.5px] font-semibold">{t("app.name")}</span>
      </header>

      <MessageList
        messages={messages ?? []}
        streamingText={stream.buffer}
        streaming={stream.streaming}
        truncated={stream.truncated}
        reconnecting={stream.reconnecting}
        siblingsOf={siblingsOf}
        onSelectBranch={(messageId) => switchBranch.mutate(messageId)}
      />

      <Composer
        value={draft}
        onChange={(text) => setDraft(id, text)}
        onSend={send}
        onStop={() => void stream.stop()}
        streaming={stream.streaming}
        attachments={uploads.items}
        onAttach={(file) => void uploads.upload(file)}
        onRemoveAttachment={uploads.remove}
        storageDown={uploads.storageDown}
      />
    </>
  );
}
