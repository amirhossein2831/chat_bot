import { useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";

import type { Message } from "./chatApi";
import { BranchSwitcher } from "./BranchSwitcher";
import { Markdown } from "./Markdown";

type Props = {
  messages: Message[];
  streamingText: string;
  streamingId: string | null;
  awaiting: boolean;
  truncated: boolean;
  reconnecting: boolean;
  failed: boolean;
  siblingsOf: (message: Message) => string[];
  onSelectBranch: (messageId: string) => void;
};

export function MessageList({
  messages,
  streamingText,
  streamingId,
  awaiting,
  truncated,
  reconnecting,
  failed,
  siblingsOf,
  onSelectBranch,
}: Props) {
  const { t } = useTranslation();
  const bottom = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottom.current?.scrollIntoView({ block: "end", behavior: "smooth" });
  }, [messages.length, streamingText, awaiting]);

  return (
    <div className="flex flex-1 justify-center overflow-y-auto px-4 pt-6 md:px-7">
      <div className="flex w-full max-w-[720px] flex-col gap-6 pb-4">
        {truncated && <Notice>{t("chat.truncated")}</Notice>}
        {reconnecting && <Notice>{t("chat.reconnecting")}</Notice>}
        {failed && <Notice tone="error">{t("chat.failed")}</Notice>}

        {/* Keyed by position, not id: the optimistic user turn and the server
            copy that replaces it share a slot, so the bubble is never
            remounted and never replays its entry animation. */}
        {messages.map((message, index) =>
          message.role === "user" ? (
            <UserTurn key={index} message={message} />
          ) : (
            <AssistantTurn
              key={index}
              message={message}
              // Only the message this stream was told it is writing into shows
              // the buffer. Assuming it is the last one blanks the previous
              // answer for as long as a new send takes to be confirmed.
              text={message.id === streamingId ? streamingText : message.content}
              streaming={message.id === streamingId}
              siblings={siblingsOf(message)}
              onSelectBranch={onSelectBranch}
            />
          ),
        )}

        {awaiting && <Thinking />}

        <div ref={bottom} />
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
        "flex items-center gap-2.5 self-center rounded-full px-3.5 py-1.5 text-xs",
        tone === "error"
          ? "bg-red-500/10 text-red-600 dark:text-red-400"
          : "bg-line/60 text-muted",
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

function UserTurn({ message }: { message: Message }) {
  return (
    <div dir="ltr" className="turn-in flex justify-end">
      <div className="max-w-[85%] rounded-[16px] rounded-ee-[5px] bg-accent-soft px-4 py-2.5 text-[15px]">
        <div dir="auto" className="whitespace-pre-wrap">
          {message.content}
        </div>
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
};

function AssistantTurn({ message, text, streaming, siblings, onSelectBranch }: AssistantProps) {
  const { t } = useTranslation();

  // The reply carries its own direction: a Persian answer reads from the
  // right, an English one from the left, whatever the interface language is.
  return (
    <div dir="auto" className="turn-in flex flex-col gap-3">
      {streaming && !text ? <Thinking /> : <Markdown>{text}</Markdown>}

      <div className="flex items-center gap-3.5">
        <BranchSwitcher siblings={siblings} current={message.id} onSelect={onSelectBranch} />

        {streaming && (
          <span className="flex items-center gap-1.5 text-xs text-muted">
            <span className="size-1.5 animate-pulse rounded-full bg-accent" />
            {t("chat.writing")}
          </span>
        )}
      </div>
    </div>
  );
}
