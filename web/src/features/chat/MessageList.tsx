import { useCallback, useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";

import type { Message } from "./chatApi";
import { BranchSwitcher } from "./BranchSwitcher";
import { Markdown } from "./Markdown";
import { MessageActions } from "./MessageActions";
import { MessageEditor } from "./MessageEditor";

type Props = {
  messages: Message[];
  streamingText: string;
  streaming: boolean;
  streamingId: string | null;
  truncated: boolean;
  reconnecting: boolean;
  failed: boolean;
  siblingsOf: (message: Message) => string[];
  onSelectBranch: (messageId: string) => void;
  onRegenerate: (messageId: string) => void;
  onEdit: (messageId: string, content: string) => void;
};

export function MessageList({
  messages,
  streamingText,
  streaming,
  streamingId,
  truncated,
  reconnecting,
  failed,
  siblingsOf,
  onSelectBranch,
  onRegenerate,
  onEdit,
}: Props) {
  const { t } = useTranslation();
  const viewport = useRef<HTMLDivElement>(null);
  const content = useRef<HTMLDivElement>(null);
  const stick = useRef(true);
  const lastTop = useRef(0);
  const lastHeight = useRef(0);

  // Instant, never smooth: an animated scroll keeps travelling toward the
  // height it was started with, so an answer that grows mid-flight lands short
  // by whatever it added. The turns fade in on their own; the scroll should
  // simply stay at the bottom.
  const pin = useCallback(() => {
    const element = viewport.current;
    if (!element || !stick.current) return;
    element.scrollTop = element.scrollHeight;
  }, []);

  // The answer keeps growing after the last token: the stored copy replaces the
  // buffer and re-renders taller. Watching the size covers every one of those
  // steps, where watching the message list misses the ones that add no row.
  useEffect(() => {
    const element = content.current;
    if (!element) return;

    const observer = new ResizeObserver(pin);
    observer.observe(element);
    return () => observer.disconnect();
  }, [pin]);

  useEffect(() => {
    stick.current = true;
    pin();
  }, [messages.length, pin]);

  return (
    <div
      ref={viewport}
      // Only a deliberate upward scroll gives up the bottom, and only when the
      // content kept its size: an answer that shrinks for a frame, as the
      // stored copy replaces the buffer, has its scrollTop clamped by the
      // browser, and that is indistinguishable from a user scrolling up.
      onScroll={(event) => {
        const element = event.currentTarget;
        const resized = element.scrollHeight !== lastHeight.current;

        if (!resized && element.scrollTop < lastTop.current - 1) stick.current = false;
        if (element.scrollHeight - element.scrollTop - element.clientHeight < 64) {
          stick.current = true;
        }

        lastTop.current = element.scrollTop;
        lastHeight.current = element.scrollHeight;
      }}
      // A column, not a row: stretched by a row container the thread would be
      // locked to the viewport height and merely overflow it, so nothing
      // observing it would ever see the answer grow.
      className="flex min-h-0 flex-1 flex-col items-center overflow-y-auto px-4 pt-4 md:px-7"
    >
      <div ref={content} className="flex w-full max-w-[736px] flex-col gap-7 pb-6">
        {truncated && <Notice>{t("chat.truncated")}</Notice>}
        {reconnecting && <Notice>{t("chat.reconnecting")}</Notice>}
        {failed && <Notice tone="error">{t("chat.failed")}</Notice>}

        {/* Keyed by position, not id: the optimistic user turn and the server
            copy that replaces it share a slot, so the bubble is never
            remounted and never replays its entry animation. */}
        {messages.map((message, index) =>
          message.role === "user" ? (
            <UserTurn
              key={index}
              message={message}
              siblings={siblingsOf(message)}
              onSelectBranch={onSelectBranch}
              onEdit={onEdit}
            />
          ) : (
            <AssistantTurn
              key={index}
              message={message}
              // Only the message this stream was told it is writing into shows
              // the buffer. Assuming it is the last one blanks the previous
              // answer for as long as a new send takes to be confirmed.
              // Both conditions matter: the id alone keeps pointing at this
              // message after the stream ends, and would leave it rendering an
              // empty buffer for good.
              text={streaming && message.id === streamingId ? streamingText : message.content}
              streaming={streaming && message.id === streamingId}
              siblings={siblingsOf(message)}
              onSelectBranch={onSelectBranch}
              onRegenerate={onRegenerate}
            />
          ),
        )}

      </div>
    </div>
  );
}

function Thinking() {
  return (
    <div dir="auto" className="flex items-center gap-1.5 py-1">
      {[0, 1, 2].map((index) => (
        <span
          key={index}
          className="thinking-dot size-1.5 rounded-full bg-muted"
          style={{ animationDelay: `${index * 160}ms` }}
        />
      ))}
    </div>
  );
}

function Notice({ children, tone }: { children: React.ReactNode; tone?: "error" }) {
  return (
    <div
      className={[
        "flex items-center gap-2 self-center rounded-full px-3 py-1.5 text-xs",
        tone === "error"
          ? "bg-red-500/10 text-red-600 dark:text-red-400"
          : "bg-raised text-muted",
      ].join(" ")}
    >
      <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
        <circle cx="12" cy="12" r="9" />
        <path d="M12 8v4.5M12 16h.01" />
      </svg>
      {children}
    </div>
  );
}

type UserProps = {
  message: Message;
  siblings: string[];
  onSelectBranch: (messageId: string) => void;
  onEdit: (messageId: string, content: string) => void;
};

function UserTurn({ message, siblings, onSelectBranch, onEdit }: UserProps) {
  const [editing, setEditing] = useState(false);

  if (editing) {
    return (
      <MessageEditor
        initial={message.content}
        onCancel={() => setEditing(false)}
        onSave={(content) => {
          setEditing(false);
          onEdit(message.id, content);
        }}
      />
    );
  }

  return (
    <div dir="ltr" className="group flex flex-col items-end gap-1">
      {/* Neutral, not accent-tinted: reserving colour for actions is what
          keeps a long thread calm to read. */}
      <div className="max-w-[82%] rounded-[var(--radius-bubble)] rounded-ee-[7px] bg-raised px-4 py-2.5 text-[15.5px] leading-[1.7]">
        <div dir="auto" className="whitespace-pre-wrap">
          {message.content}
        </div>
      </div>

      <div className="flex items-center gap-2">
        <BranchSwitcher siblings={siblings} current={message.id} onSelect={onSelectBranch} />
        {/* The optimistic copy has no server id yet, so there is nothing to
            branch from until it is confirmed. */}
        <MessageActions
          text={message.content}
          onEdit={message.id === "pending" ? undefined : () => setEditing(true)}
        />
      </div>
    </div>
  );
}

type AssistantProps = {
  message: Message;
  text: string;
  streaming: boolean;
  siblings: string[];
  onSelectBranch: (messageId: string) => void;
  onRegenerate: (messageId: string) => void;
};

function AssistantTurn({
  message,
  text,
  streaming,
  siblings,
  onSelectBranch,
  onRegenerate,
}: AssistantProps) {
  // The reply carries its own direction: a Persian answer reads from the
  // right, an English one from the left, whatever the interface language is.
  return (
    <div
      dir="auto"
      className={["group flex flex-col gap-2 text-[15.5px]", streaming && text ? "caret" : ""].join(" ")}
    >
      {streaming && !text ? <Thinking /> : <Markdown>{text}</Markdown>}

      <div className="flex items-center gap-2">
        <BranchSwitcher siblings={siblings} current={message.id} onSelect={onSelectBranch} />

        {/* An answer still being written has nothing settled to copy, and
            regenerating it would race the generation it is replacing. */}
        {!streaming && text && (
          <MessageActions text={text} onRegenerate={() => onRegenerate(message.id)} />
        )}
      </div>
    </div>
  );
}
