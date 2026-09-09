import { useState } from "react";
import { useTranslation } from "react-i18next";

type Props = {
  text: string;
  onRegenerate?: () => void;
  onEdit?: () => void;
};

// Shown on hover on a pointer device and always on touch, where there is no
// hover to reveal anything.
export function MessageActions({ text, onRegenerate, onEdit }: Props) {
  const { t } = useTranslation();
  const [copied, setCopied] = useState(false);

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 1600);
    } catch {
      setCopied(false);
    }
  };

  return (
    <div className="flex items-center gap-0.5 text-muted transition-opacity lg:opacity-0 lg:group-hover:opacity-100 lg:group-focus-within:opacity-100">
      <Action label={t(copied ? "chat.copied" : "chat.copy")} onClick={() => void copy()}>
        {copied ? (
          <path d="m4 12.5 5 5L20 6.5" />
        ) : (
          <>
            <rect x="9" y="9" width="11" height="11" rx="2.5" />
            <path d="M5 15V5.5A1.5 1.5 0 0 1 6.5 4H15" />
          </>
        )}
      </Action>

      {onRegenerate && (
        <Action label={t("chat.regenerate")} onClick={onRegenerate}>
          <path d="M3 12a9 9 0 1 0 3-6.7L3 8" />
          <path d="M3 3v5h5" />
        </Action>
      )}

      {onEdit && (
        <Action label={t("chat.edit")} onClick={onEdit}>
          <path d="M4 20h4L19 9a2.8 2.8 0 0 0-4-4L4 16z" />
        </Action>
      )}
    </div>
  );
}

function Action({
  label,
  onClick,
  children,
}: {
  label: string;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={label}
      title={label}
      className="flex size-7 items-center justify-center rounded-lg transition hover:bg-raised hover:text-ink"
    >
      <svg
        width="14"
        height="14"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.9"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        {children}
      </svg>
    </button>
  );
}
