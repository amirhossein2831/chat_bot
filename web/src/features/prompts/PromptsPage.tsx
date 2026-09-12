import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import { useUiStore } from "../../app/store";
import { useDefaultModel } from "../chat/useModels";
import { useCreateConversation } from "../conversations/useConversations";
import { usePrompts } from "./usePrompts";

export function PromptsPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();

  const locale = useUiStore((state) => state.locale);
  const setDraft = useUiStore((state) => state.setDraft);

  const [category, setCategory] = useState("");
  const { data: prompts } = usePrompts(locale, "");
  const create = useCreateConversation();
  const model = useDefaultModel();

  const categories = useMemo(
    () => [...new Set((prompts ?? []).map((prompt) => prompt.category))],
    [prompts],
  );

  const shown = category
    ? (prompts ?? []).filter((prompt) => prompt.category === category)
    : (prompts ?? []);

  // A prompt is a starting point, not a message: it lands in the composer so
  // the user fills the blanks before sending.
  const start = async (body: string, title: string) => {
    const conversation = await create.mutateAsync({ title, model });
    setDraft(conversation.id, body);
    navigate(`/c/${conversation.id}`);
  };

  return (
    <>
      <header className="flex h-14 shrink-0 items-center gap-3 border-b border-line px-4 md:px-10">
        <button
          type="button"
          onClick={() => navigate(-1)}
          className="flex h-11 items-center gap-1.5 rounded-lg px-2 text-[13px] text-muted md:h-8"
        >
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="m15 6 6 6-6 6" />
          </svg>
          {t("nav.back")}
        </button>
        <span className="text-[15px] font-semibold">{t("prompts.title")}</span>
      </header>

      <div className="flex flex-1 justify-center overflow-y-auto px-4 py-8 md:px-10">
        <div className="flex w-full max-w-[940px] flex-col gap-6">
          <p className="m-0 max-w-[62ch] text-[14.5px] text-muted">{t("prompts.subtitle")}</p>

          <div className="flex flex-wrap items-center gap-2">
            <Chip active={category === ""} onClick={() => setCategory("")}>
              {t("prompts.all")}
            </Chip>
            {categories.map((name) => (
              <Chip key={name} active={category === name} onClick={() => setCategory(name)}>
                {name}
              </Chip>
            ))}
          </div>

          <div className="grid grid-cols-1 gap-3.5 md:grid-cols-2 lg:grid-cols-3">
            {shown.map((prompt) => (
              <button
                key={prompt.id}
                type="button"
                onClick={() => void start(prompt.body, prompt.title)}
                className="flex flex-col gap-2.5 rounded-[13px] border border-line bg-surface p-[18px] text-start hover:border-accent/40"
              >
                <span className="text-[11.5px] font-semibold text-accent">{prompt.category}</span>
                <span className="text-[15px] font-semibold">{prompt.title}</span>
                <span className="line-clamp-3 text-[13.5px] leading-loose text-muted">
                  {prompt.body}
                </span>
              </button>
            ))}
          </div>
        </div>
      </div>
    </>
  );
}

type ChipProps = { active: boolean; onClick: () => void; children: React.ReactNode };

function Chip({ active, onClick, children }: ChipProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={[
        "h-11 rounded-full px-3.5 text-[13px] md:h-8",
        active
          ? "bg-accent font-medium text-white"
          : "border border-line bg-surface text-ink hover:border-accent/40",
      ].join(" ")}
    >
      {children}
    </button>
  );
}
