import { useCallback, useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { useLocation, useNavigate, useParams } from "react-router-dom";

import { useUiStore } from "../../app/store";
import { useCreateConversation } from "../conversations/useConversations";
import { useUpload } from "../files/useUpload";
import type { Message } from "./chatApi";
import { Composer } from "./Composer";
import { MessageList } from "./MessageList";
import { ModelPicker } from "./ModelPicker";
import { useChatStream } from "./useChatStream";
import { useDefaultModel } from "./useModels";
import { useSwitchBranch, useThread } from "./useThread";

type PendingSend = { pending?: string; model?: string };

function optimisticTurn(content: string): Message {
  return {
    id: "pending",
    parentId: null,
    role: "user",
    content,
    status: "complete",
    model: "",
    tokensIn: 0,
    tokensOut: 0,
    error: null,
    createdAt: new Date().toISOString(),
  };
}

function hasStreamingTurn(messages: Message[]): boolean {
  const last = messages[messages.length - 1];
  return last?.role === "assistant";
}

export function ChatPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const { id = "" } = useParams();

  const drafts = useUiStore((state) => state.drafts);
  const setDraft = useUiStore((state) => state.setDraft);
  const setSidebarOpen = useUiStore((state) => state.setSidebarOpen);

  const fallbackModel = useDefaultModel();
  const [model, setModel] = useState("");

  const { data: messages, refetch } = useThread(id);
  const switchBranch = useSwitchBranch(id);
  const create = useCreateConversation();
  const uploads = useUpload();

  const onFinished = useCallback(() => {
    void refetch();
  }, [refetch]);

  const stream = useChatStream(onFinished, onFinished);

  const chosenModel = model || fallbackModel;

  // Reset and pending-send must happen in one effect and in this order.
  // Split across two, the reset aborted the send that had just started.
  useEffect(() => {
    stream.reset();

    const state = location.state as PendingSend | null;
    if (id && state?.pending) {
      navigate(location.pathname, { replace: true, state: null });
      void stream.send(id, state.pending, state.model);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id]);

  useEffect(() => {
    const last = messages?.[messages.length - 1];
    if (last?.role === "assistant" && last.status === "streaming" && !stream.streaming) {
      stream.resume(last.id);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [messages]);

  const draftKey = id || "new";
  const draft = drafts[draftKey] ?? "";

  const send = async () => {
    const content = draft.trim();
    if (!content) return;

    setDraft(draftKey, "");
    uploads.clear();

    if (!id) {
      const conversation = await create.mutateAsync({ title: content.slice(0, 60), model: chosenModel });
      navigate(`/c/${conversation.id}`, { state: { pending: content, model: chosenModel } });
      return;
    }

    void stream.send(id, content, chosenModel);
  };

  const siblingsOf = useMemo(() => {
    const byParent = new Map<string | null, string[]>();
    for (const message of messages ?? []) {
      const existing = byParent.get(message.parentId) ?? [];
      byParent.set(message.parentId, [...existing, message.id]);
    }
    return (message: Message) => byParent.get(message.parentId) ?? [];
  }, [messages]);

  // The optimistic turn is dropped as soon as the server's own copy of it
  // arrives, so the same message is never on screen twice.
  const shown = useMemo(() => {
    const thread = messages ?? [];
    if (stream.sent === null) return thread;

    const lastUser = [...thread].reverse().find((message) => message.role === "user");
    if (lastUser?.content === stream.sent) return thread;

    return [...thread, optimisticTurn(stream.sent)];
  }, [messages, stream.sent]);

  const empty = !id || (shown.length === 0 && !stream.streaming);
  const awaiting = stream.streaming && !stream.buffer && !hasStreamingTurn(shown);

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

        <span className="min-w-0 truncate text-[14.5px] font-semibold">{t("app.name")}</span>

        <div className="ms-auto">
          <ModelPicker value={chosenModel} onChange={setModel} />
        </div>
      </header>

      {stream.failed && (
        <div className="flex justify-center px-4 pt-4">
          <div className="flex items-center gap-2.5 rounded-full bg-red-500/10 px-3.5 py-1.5 text-xs text-red-600 dark:text-red-400">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <circle cx="12" cy="12" r="9" />
              <path d="M12 8v4.5M12 16h.01" />
            </svg>
            {t(stream.failureCode === "model_unavailable" ? "chat.modelDown" : "chat.failed")}
          </div>
        </div>
      )}

      {empty ? (
        <div className="flex flex-1 flex-col items-center justify-center gap-3 px-6 text-center">
          <div className="flex size-12 items-center justify-center rounded-xl bg-accent-soft">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round" className="text-accent">
              <path d="M21 11.5a8.4 8.4 0 0 1-9 8.4 8.9 8.9 0 0 1-3.8-.9L3 21l1.9-5.1A8.4 8.4 0 0 1 12 3a8.4 8.4 0 0 1 9 8.5z" />
            </svg>
          </div>
          <h1 className="m-0 text-xl font-semibold">{t("empty.title")}</h1>
          <p className="m-0 max-w-[46ch] text-sm text-muted">{t("empty.subtitle")}</p>
        </div>
      ) : (
        <MessageList
          messages={shown}
          streamingText={stream.buffer}
          streaming={stream.streaming}
          awaiting={awaiting}
          truncated={stream.truncated}
          reconnecting={stream.reconnecting}
          failed={false}
          siblingsOf={siblingsOf}
          onSelectBranch={(messageId) => switchBranch.mutate(messageId)}
        />
      )}

      <Composer
        value={draft}
        onChange={(text) => setDraft(draftKey, text)}
        onSend={() => void send()}
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
