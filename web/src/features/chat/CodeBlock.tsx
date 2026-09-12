import { useState } from "react";
import { useTranslation } from "react-i18next";

type Props = {
  language: string;
  source: string;
  children: React.ReactNode;
};

export function CodeBlock({ language, source, children }: Props) {
  const { t } = useTranslation();
  const [copied, setCopied] = useState(false);

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(source);
      setCopied(true);
      setTimeout(() => setCopied(false), 1600);
    } catch {
      setCopied(false);
    }
  };

  return (
    <div dir="ltr" className="overflow-hidden rounded-[14px] bg-raised">
      <div className="flex items-center justify-between gap-2 ps-4 pe-2 py-1.5">
        <span className="font-mono text-[11.5px] tracking-wide text-muted">{language}</span>

        <button
          type="button"
          onClick={() => void copy()}
          className="flex items-center gap-1.5 rounded-lg px-2 py-1 text-[11.5px] text-muted transition hover:bg-line/70"
        >
          {copied ? (
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.4" strokeLinecap="round" strokeLinejoin="round">
              <path d="m4 12.5 5 5L20 6.5" />
            </svg>
          ) : (
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round">
              <rect x="9" y="9" width="11" height="11" rx="2.5" />
              <path d="M5 15V5.5A1.5 1.5 0 0 1 6.5 4H15" />
            </svg>
          )}
          {t(copied ? "chat.copied" : "chat.copy")}
        </button>
      </div>

      <pre className="overflow-x-auto px-4 pt-1 pb-4 font-mono text-[13px] leading-relaxed">
        {children}
      </pre>
    </div>
  );
}
