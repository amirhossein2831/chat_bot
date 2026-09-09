import { useUiStore } from "../../app/store";
import { usePrompts } from "../prompts/usePrompts";

type Props = {
  onPick: (body: string) => void;
};

const SHOWN = 4;

// The prompt library is the one thing a new conversation can offer that is
// specific to this product rather than generic encouragement.
export function Suggestions({ onPick }: Props) {
  const locale = useUiStore((state) => state.locale);
  const { data: prompts } = usePrompts(locale, "");

  const shown = (prompts ?? []).slice(0, SHOWN);
  if (shown.length === 0) return null;

  return (
    <div className="grid w-full max-w-[540px] grid-cols-1 gap-2 sm:grid-cols-2">
      {shown.map((prompt) => (
        <button
          key={prompt.id}
          type="button"
          onClick={() => onPick(prompt.body)}
          className="flex min-h-[64px] flex-col gap-1 rounded-[14px] bg-surface p-3 text-start transition hover:bg-raised"
        >
          <span dir="auto" className="text-[13.5px] font-medium">
            {prompt.title}
          </span>
          <span dir="auto" className="line-clamp-2 text-[12.5px] text-muted">
            {prompt.body}
          </span>
        </button>
      ))}
    </div>
  );
}
