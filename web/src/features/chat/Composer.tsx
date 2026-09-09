import { useLayoutEffect, useRef } from "react";
import { useTranslation } from "react-i18next";

import { useVisualViewport } from "./useVisualViewport";

type Props = {
  value: string;
  onChange: (value: string) => void;
  onSend: () => void;
  onStop: () => void;
  streaming: boolean;
};

const MAX_ROWS = 8;

export function Composer({ value, onChange, onSend, onStop, streaming }: Props) {
  const { t } = useTranslation();
  const textarea = useRef<HTMLTextAreaElement>(null);
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
        <div className="flex items-end gap-2.5 rounded-[14px] border border-line bg-surface px-3 py-2.5">
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
