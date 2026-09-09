import { useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";

import type { Message } from "./chatApi";
import { BranchSwitcher } from "./BranchSwitcher";
import { Markdown } from "./Markdown";

type Props = {
  messages: Message[];
  streamingText: string;
  streaming: boolean;
  truncated: boolean;
  reconnecting: boolean;
  failed: boolean;
  siblingsOf: (message: Message) => string[];
  onSelectBranch: (messageId: string) => void;
};

export function MessageList({
  messages,
  streamingText,
  streaming,
  truncated,
  reconnecting,
  failed,
  siblingsOf,
  onSelectBranch,
}: Props) {
  const { t } = useTranslation();
  const bottom = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottom.current?.scrollIntoView({ block: "end" });
  }, [messages.length, streamingText]);

  return (
    <div className="flex flex-1 justify-center overflow-y-auto px-4 pt-6 md:px-7">
      <div className="flex w-full max-w-[720px] flex-col gap-6 pb-4">
        {truncated && <Notice>{t("chat.truncated")}</Notice>}
        {reconnecting && <Notice>{t("chat.reconnecting")}</Notice>}
        {failed && <Notice tone="error">{t("chat.failed")}</Notice>}

        {messages.map((message) =>
          message.role === "user" ? (
            <UserTurn key={message.id} message={message} />
          ) : (
            <AssistantTurn
              key={message.id}
              message={message}
              // The last assistant message is the one being written into, so
              // while streaming its stored content is stale by design.
              text={
                streaming && message.id === messages[messages.length - 1]?.id
                  ? streamingText
                  : message.content
              }
              streaming={streaming && message.id === messages[messages.length - 1]?.id}
              siblings={siblingsOf(message)}
              onSelectBranch={onSelectBranch}
            />
          ),
        )}

        <div ref={bottom} />
      </div>
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
    <div className="flex justify-start">
      <div className="max-w-[85%] rounded-[14px] rounded-es-[4px] bg-accent-soft px-4 py-3 text-[14.5px]">
        {message.content}
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

  return (
    <div className="flex justify-end gap-3">
      <div className="flex min-w-0 max-w-[85%] flex-col items-end gap-2">
        <div className="w-full rounded-[14px] rounded-ee-[4px] border border-line bg-raised px-4 py-3">
          <Markdown>{text}</Markdown>
        </div>

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

      <div className="mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-lg bg-accent">
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.1" strokeLinecap="round" strokeLinejoin="round">
          <path d="M21 11.5a8.4 8.4 0 0 1-9 8.4 8.9 8.9 0 0 1-3.8-.9L3 21l1.9-5.1A8.4 8.4 0 0 1 12 3a8.4 8.4 0 0 1 9 8.5z" />
        </svg>
      </div>
    </div>
  );
}
