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
      className="flex justify-center px-4 pt-3 pb-5 md:px-7"
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

        <div className="flex items-end gap-2.5 rounded-[14px] border border-line bg-surface px-3 py-2.5">
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
            className="flex size-11 shrink-0 items-center justify-center rounded-lg text-muted disabled:opacity-40 md:size-9"
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
            className="max-h-[208px] flex-1 resize-none bg-transparent py-1.5 text-[14.5px] outline-none placeholder:text-muted"
          />

          <button
            type="button"
            onClick={streaming ? onStop : onSend}
            disabled={!streaming && !value.trim()}
            aria-label={streaming ? t("composer.stop") : t("composer.send")}
            className="flex size-11 shrink-0 items-center justify-center rounded-[9px] bg-accent text-white disabled:opacity-40 md:size-9"
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

        <span className="px-1 text-[11.5px] text-muted">{t("composer.hint")}</span>
      </div>
    </div>
  );
}
