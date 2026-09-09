import { useTranslation } from "react-i18next";

import { useUiStore } from "../../app/store";
import { formatNumber } from "../../lib/format";

type Props = {
  siblings: string[];
  current: string;
  onSelect: (messageId: string) => void;
};

// Regenerating creates a sibling rather than replacing the previous answer, so
// this is how the other branches stay reachable.
export function BranchSwitcher({ siblings, current, onSelect }: Props) {
  const { t } = useTranslation();
  const locale = useUiStore((state) => state.locale);

  if (siblings.length < 2) return null;

  const index = siblings.indexOf(current);

  const go = (delta: number) => {
    const next = (index + delta + siblings.length) % siblings.length;
    onSelect(siblings[next]);
  };

  return (
    <div className="flex items-center gap-0.5 rounded-lg border border-line bg-surface p-1">
      <button
        type="button"
        onClick={() => go(-1)}
        aria-label="previous"
        className="flex size-8 items-center justify-center rounded-md text-muted hover:bg-line/50"
      >
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="m9 6-6 6 6 6" />
        </svg>
      </button>

      <span className="px-1 text-xs text-muted">
        {t("chat.branch", {
          current: formatNumber(index + 1, locale),
          total: formatNumber(siblings.length, locale),
        })}
      </span>

      <button
        type="button"
        onClick={() => go(1)}
        aria-label="next"
        className="flex size-8 items-center justify-center rounded-md text-muted hover:bg-line/50"
      >
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="m15 6 6 6-6 6" />
        </svg>
      </button>
    </div>
  );
}
