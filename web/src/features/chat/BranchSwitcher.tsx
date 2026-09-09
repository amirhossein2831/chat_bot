import { useTranslation } from "react-i18next";

import { useUiStore } from "../../app/store";
import { formatNumber } from "../../lib/format";
import { IconButton } from "./IconButton";

type Props = {
  siblings: string[];
  current: string;
  onSelect: (messageId: string) => void;
};

// Regenerating and editing create siblings rather than replacing what was
// there, so this is how the other branches stay reachable.
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
    <div className="flex items-center text-muted">
      <IconButton mirror label={t("chat.previousBranch")} onClick={() => go(-1)}>
        <path d="m14 6-6 6 6 6" />
      </IconButton>

      <span className="px-0.5 text-[11.5px] tabular-nums">
        {formatNumber(index + 1, locale)}/{formatNumber(siblings.length, locale)}
      </span>

      <IconButton mirror label={t("chat.nextBranch")} onClick={() => go(1)}>
        <path d="m10 6 6 6-6 6" />
      </IconButton>
    </div>
  );
}
