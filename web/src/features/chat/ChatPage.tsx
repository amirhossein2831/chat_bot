import { useCallback, useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useLocation, useNavigate, useParams } from "react-router-dom";

import { useUiStore } from "../../app/store";
import { useConversations, useCreateConversation } from "../conversations/useConversations";
import { useUpload } from "../files/useUpload";
import type { Message } from "./chatApi";
import { Composer } from "./Composer";
import { MessageList } from "./MessageList";
import { ModelPicker } from "./ModelPicker";
import { useChatStream } from "./useChatStream";
import { useDefaultModel } from "./useModels";
import { Suggestions } from "./Suggestions";
import { useSwitchBranch, useThread } from "./useThread";

type PendingSend = { pending?: string; model?: string; attachments?: string[] };

// The waiting reply is a turn in the list rather than a separate indicator, so
// the dots and the answer that replaces them are the same element and nothing
// jumps into place when the first token lands.
const pendingAssistantId = "pending-assistant";

// Every one of these reaches the user as "something went wrong" unless it is
// named here, which hides the two failures they can actually act on.
// Retrying a limit or a permission refusal cannot succeed, so the button is
// not offered for them.
function pendingAssistant(at: number): Message {
  return {
    id: pendingAssistantId,
    parentId: "pending",
    role: "assistant",
    content: "",
    status: "streaming",
    model: "",
    tokensIn: 0,
    tokensOut: 0,
    error: null,
    siblings: [],
    createdAt: new Date(at).toISOString(),
  };
}

function retryable(code: string | null): boolean {
  return code !== "quota_exceeded" && code !== "forbidden";
}

const failureMessages: Record<string, string> = {
  model_unavailable: "chat.modelDown",
  quota_exceeded: "chat.quotaExceeded",
  forbidden: "chat.notAllowed",
  disconnected: "chat.disconnected",
};

function optimisticTurn(content: string, at: number): Message {
  return {
    id: "pending",
    parentId: "pending",
    role: "user",
    content,
    status: "complete",
    model: "",
    tokensIn: 0,
    tokensOut: 0,
    error: null,
    siblings: [],
    createdAt: new Date(at).toISOString(),
  };
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

  const { data: messages, refetch, dataUpdatedAt } = useThread(id);
  const { data: conversations } = useConversations("");
  const conversation = conversations?.find((item) => item.id === id);
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
  //
  // Entering twice for one conversation must do nothing the second time: the
  // reset would abort the send the first pass had just begun, and the send
  // would run again, asking for two answers to one question.
  const entered = useRef<string | null>(null);

  useEffect(() => {
    if (entered.current === id) return;
    entered.current = id;

    stream.reset();

    const state = location.state as PendingSend | null;
    if (id && state?.pending) {
      navigate(location.pathname, { replace: true, state: null });
      void stream.send(id, state.pending, state.model, undefined, state.attachments);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id]);

  useEffect(() => {
    const last = messages?.[messages.length - 1];
    if (
      last?.role === "assistant" &&
      last.status === "streaming" &&
      !stream.streaming &&
      last.id !== stream.messageId
    ) {
      stream.resume(last.id);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [messages]);

  const draftKey = id || "new";
  const draft = drafts[draftKey] ?? "";

  const retry = () => {
    const last = shown[shown.length - 1];
    if (last?.role === "assistant" && last.status === "failed") {
      void stream.regenerate(last.id, chosenModel);
      return;
    }
    stream.retry();
  };

  const send = async () => {
    const content = draft.trim();
    if (!content) return;

    // Only documents that finished extracting carry text; one still being read
    // would reach the model as nothing at all.
    const attachments = uploads.items
      .filter((item) => item.status === "done")
      .map((item) => item.id);

    setDraft(draftKey, "");
    uploads.clear();

    if (!id) {
      const conversation = await create.mutateAsync({ title: content.slice(0, 60), model: chosenModel });
      navigate(`/c/${conversation.id}`, {
        state: { pending: content, model: chosenModel, attachments },
      });
      return;
    }

    void stream.send(id, content, chosenModel, undefined, attachments);
  };

  // The branches beside this one are named by the server: a thread is a single
  // path, so nothing here could work them out on its own.
  const siblingsOf = (message: Message) => message.siblings;

  const thread = messages ?? [];

  // Whether the server's own copy of this turn is in hand, judged by the id it
  // named rather than by content or timing. Content would hide a message
  // identical to the one before it, and timing counts the first fetch of a new
  // conversation, which lands before the turn exists and leaves the thread
  // empty for as long as the model takes to answer.
  const arrived = stream.messageId !== null && thread.some((m) => m.id === stream.messageId);
  const abandoned = !stream.streaming && stream.sentAt !== null && dataUpdatedAt > stream.sentAt;
  const confirmed = arrived || abandoned;
  const shown =
    stream.sent === null || stream.sentAt === null || confirmed
      ? thread
      : [...thread, optimisticTurn(stream.sent, stream.sentAt), pendingAssistant(stream.sentAt)];

  const empty = !id || (shown.length === 0 && !stream.streaming);

  return (
    <>
      <header className="flex h-[58px] shrink-0 items-center gap-3 px-3 md:px-5">
        <button
          type="button"
          onClick={() => setSidebarOpen(true)}
          aria-label={t("nav.newChat")}
          className="flex size-10 items-center justify-center rounded-[12px] text-muted transition hover:bg-surface lg:hidden"
        >
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
            <path d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>

        {/* The sidebar already carries the product name; repeating it here
            costs the one line that says which conversation this is. */}
        {/* bdi, not dir="auto": the title must isolate its own direction while
            still being aligned by the interface's. */}
        <span className="min-w-0 flex-1 truncate text-start text-[14px] font-medium text-muted">
          <bdi>{conversation?.title}</bdi>
        </span>

        <ModelPicker value={chosenModel} onChange={setModel} />
      </header>

      {stream.failed && (
        <div className="flex justify-center px-4 pt-4">
          <div className="flex items-center gap-2.5 rounded-full bg-red-500/10 px-3.5 py-1.5 text-xs text-red-600 dark:text-red-400">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <circle cx="12" cy="12" r="9" />
              <path d="M12 8v4.5M12 16h.01" />
            </svg>
            {t(failureMessages[stream.failureCode ?? ""] ?? "chat.failed")}
            {retryable(stream.failureCode) && (
              <button
                type="button"
                onClick={retry}
                className="flex items-center gap-1 font-medium underline underline-offset-2"
              >
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M3 12a9 9 0 1 0 3-6.7L3 8" />
                  <path d="M3 3v5h5" />
                </svg>
                {t("chat.retry")}
              </button>
            )}
          </div>
        </div>
      )}

      {empty ? (
        <div className="flex flex-1 flex-col items-center justify-center gap-6 px-6 pb-10">
          <div className="flex flex-col items-center gap-2.5 text-center">
            <h1 className="m-0 text-[26px] font-semibold tracking-tight">{t("empty.title")}</h1>
            <p className="m-0 max-w-[46ch] text-[14.5px] text-muted">{t("empty.subtitle")}</p>
          </div>

          <Suggestions onPick={(body) => setDraft(draftKey, body)} />
        </div>
      ) : (
        <MessageList
          messages={shown}
          streamingText={stream.buffer}
          streaming={stream.streaming}
          streamingId={stream.messageId ?? pendingAssistantId}
          truncated={stream.truncated}
          reconnecting={stream.reconnecting}
          failed={false}
          siblingsOf={siblingsOf}
          onSelectBranch={(messageId) => switchBranch.mutate(messageId)}
          onRegenerate={(messageId) => void stream.regenerate(messageId, chosenModel)}
          onEdit={(messageId, content) => void stream.send(id, content, chosenModel, messageId)}
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
