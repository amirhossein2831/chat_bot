import { useLayoutEffect, useRef } from "react";
import { useTranslation } from "react-i18next";

import { AttachmentChip } from "../files/AttachmentChip";
import type { UploadItem } from "../files/useUpload";
import { useVisualViewport } from "./useVisualViewport";

type Props = {
  value: string;
  onChange: (value: string) => void;
  onSend: () => void;
  onStop: () => void;
  streaming: boolean;
  attachments: UploadItem[];
  onAttach: (file: File) => void;
  onRemoveAttachment: (id: string) => void;
  storageDown: boolean;
};

const MAX_ROWS = 8;

export function Composer({
  value,
  onChange,
  onSend,
  onStop,
  streaming,
  attachments,
  onAttach,
  onRemoveAttachment,
  storageDown,
}: Props) {
  const { t } = useTranslation();
  const textarea = useRef<HTMLTextAreaElement>(null);
  const fileInput = useRef<HTMLInputElement>(null);
  const keyboardInset = useVisualViewport();

  useLayoutEffect(() => {
    const element = textarea.current;
    if (!element) return;

    element.style.height = "auto";
    const lineHeight = 26;
    element.style.height = `${Math.min(element.scrollHeight, lineHeight * MAX_ROWS)}px`;
  }, [value]);

  const onKeyDown = (event: React.KeyboardEvent<HTMLTextAreaElement>) => {
    // Without the composing guard, Enter fires mid-IME and sends half a
    // Persian word.
    if (event.key !== "Enter" || event.shiftKey || event.nativeEvent.isComposing) return;

    event.preventDefault();
    if (!streaming && value.trim()) onSend();
  };

  return (
    <div
      className="flex shrink-0 justify-center px-4 pt-2 pb-4 md:px-7"
      style={{ paddingBottom: keyboardInset ? keyboardInset + 12 : undefined }}
    >
      <div className="flex w-full max-w-[720px] flex-col gap-2">
        {attachments.length > 0 && (
          <div className="flex flex-wrap gap-2">
            {attachments.map((item) => (
              <AttachmentChip key={item.id} item={item} onRemove={onRemoveAttachment} />
            ))}
          </div>
        )}

        {/* A raised card rather than a bordered box: the thread scrolls up to
            it and the shadow separates the two without drawing a line. */}
        <div className="flex items-end gap-1.5 rounded-[var(--radius-card)] bg-paper p-2 shadow-[var(--shadow-lift)] transition-shadow focus-within:shadow-[var(--shadow-lift),0_0_0_2px_var(--color-accent-soft)]">
          <input
            ref={fileInput}
            type="file"
            hidden
            onChange={(event) => {
              const file = event.target.files?.[0];
              if (file) onAttach(file);
              event.target.value = "";
            }}
          />

          <button
            type="button"
            onClick={() => fileInput.current?.click()}
            disabled={storageDown}
            aria-label={t("files.attach")}
            title={storageDown ? t("files.storageDown") : t("files.attach")}
            className="flex size-9 shrink-0 items-center justify-center rounded-[10px] text-muted transition hover:bg-raised disabled:opacity-40"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M21 11.5 12.5 20a5 5 0 0 1-7-7l8.5-8.5a3.4 3.4 0 0 1 4.8 4.8l-8.5 8.5a1.8 1.8 0 0 1-2.5-2.5l7.8-7.8" />
            </svg>
          </button>

          <textarea
            ref={textarea}
            rows={1}
            value={value}
            onChange={(event) => onChange(event.target.value)}
            onKeyDown={onKeyDown}
            placeholder={t("composer.placeholder")}
            className="max-h-[208px] flex-1 resize-none bg-transparent px-1 py-2 text-[15px] leading-[26px] outline-none placeholder:text-muted"
          />

          <button
            type="button"
            onClick={streaming ? onStop : onSend}
            disabled={!streaming && !value.trim()}
            aria-label={streaming ? t("composer.stop") : t("composer.send")}
            className="flex size-9 shrink-0 items-center justify-center rounded-full bg-accent text-white transition disabled:opacity-30"
          >
            {streaming ? (
              <svg width="15" height="15" viewBox="0 0 24 24" fill="currentColor">
                <rect x="6" y="6" width="12" height="12" rx="2" />
              </svg>
            ) : (
              <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.1" strokeLinecap="round" strokeLinejoin="round">
                <path d="M20 12H5" />
                <path d="m12 5-7 7 7 7" />
              </svg>
            )}
          </button>
        </div>

        <span className="px-2 text-center text-[11.5px] text-muted">{t("composer.hint")}</span>
      </div>
    </div>
  );
}
