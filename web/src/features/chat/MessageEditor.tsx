import { useLayoutEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";

type Props = {
  initial: string;
  onSave: (content: string) => void;
  onCancel: () => void;
};

export function MessageEditor({ initial, onSave, onCancel }: Props) {
  const { t } = useTranslation();
  const [value, setValue] = useState(initial);
  const textarea = useRef<HTMLTextAreaElement>(null);

  useLayoutEffect(() => {
    const element = textarea.current;
    if (!element) return;

    element.style.height = "auto";
    element.style.height = `${element.scrollHeight}px`;
    element.setSelectionRange(element.value.length, element.value.length);
  }, [value]);

  const save = () => {
    const next = value.trim();
    if (next && next !== initial) onSave(next);
    else onCancel();
  };

  return (
    <div dir="ltr" className="flex flex-col items-end gap-2">
      <div className="w-[82%] rounded-[var(--radius-bubble)] bg-raised px-4 py-2.5">
        <textarea
          ref={textarea}
          autoFocus
          dir="auto"
          value={value}
          onChange={(event) => setValue(event.target.value)}
          onKeyDown={(event) => {
            if (event.key === "Escape") onCancel();
            if (event.key === "Enter" && !event.shiftKey && !event.nativeEvent.isComposing) {
              event.preventDefault();
              save();
            }
          }}
          className="w-full resize-none bg-transparent text-[15.5px] leading-[1.7] outline-none"
        />
      </div>

      <div className="flex items-center gap-2">
        <button
          type="button"
          onClick={onCancel}
          className="rounded-[10px] px-3 py-1.5 text-[13px] text-muted transition hover:bg-raised"
        >
          {t("chat.cancel")}
        </button>
        <button
          type="button"
          onClick={save}
          className="rounded-[10px] bg-accent px-3 py-1.5 text-[13px] font-medium text-white"
        >
          {t("chat.save")}
        </button>
      </div>
    </div>
  );
}
