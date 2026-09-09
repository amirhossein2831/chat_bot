import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import { useUiStore } from "../../app/store";
import { formatNumber, formatRelative } from "../../lib/format";
import { useAddFact, useDeleteFact, useFacts, useInstructions, useSaveInstructions } from "./useMemory";

const MAX_INSTRUCTIONS = 1000;

export function MemoryPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const locale = useUiStore((state) => state.locale);

  const { data: stored } = useInstructions();
  const { data: facts } = useFacts();
  const save = useSaveInstructions();
  const add = useAddFact();
  const remove = useDeleteFact();

  // Local state holds only an unsaved edit. Copying the server value into
  // state with an effect would fight the query cache on every refetch.
  const [edited, setEdited] = useState<string | null>(null);
  const [newFact, setNewFact] = useState("");

  const instructions = edited ?? stored ?? "";
  const dirty = edited !== null && edited !== stored;

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
        <span className="text-[15px] font-semibold">{t("memory.title")}</span>
      </header>

      <div className="flex flex-1 justify-center overflow-y-auto px-4 py-8 md:px-10">
        <div className="flex w-full max-w-[720px] flex-col gap-8">

          <section className="flex flex-col gap-3">
            <div className="flex flex-col gap-1">
              <h2 className="m-0 text-[17px] font-semibold">{t("memory.instructions")}</h2>
              <p className="m-0 text-[13.5px] text-muted">{t("memory.instructionsHelp")}</p>
            </div>

            <textarea
              value={instructions}
              maxLength={MAX_INSTRUCTIONS}
              onChange={(event) => setEdited(event.target.value)}
              rows={3}
              className="resize-none rounded-xl border border-line bg-surface px-4 py-3.5 text-[14.5px] outline-none focus:border-accent"
            />

            <div className="flex items-center gap-2.5">
              <button
                type="button"
                onClick={() => save.mutate(instructions, { onSuccess: () => setEdited(null) })}
                disabled={save.isPending || !dirty}
                className="h-11 rounded-[9px] bg-accent px-4 text-[13.5px] font-medium text-white disabled:opacity-50 md:h-9"
              >
                {t("memory.save")}
              </button>
              <span className="text-[12.5px] text-muted">
                {formatNumber(instructions.length, locale)} / {formatNumber(MAX_INSTRUCTIONS, locale)}
              </span>
            </div>
          </section>

          <div className="h-px bg-line" />

          <section className="flex flex-col gap-3.5">
            <div className="flex flex-col gap-1">
              <h2 className="m-0 text-[17px] font-semibold">{t("memory.facts")}</h2>
              <p className="m-0 max-w-[58ch] text-[13.5px] text-muted">{t("memory.factsHelp")}</p>
            </div>

            <div className="flex gap-2">
              <input
                value={newFact}
                onChange={(event) => setNewFact(event.target.value)}
                maxLength={200}
                className="h-11 flex-1 rounded-[9px] border border-line bg-surface px-3 text-[14px] outline-none focus:border-accent md:h-9"
              />
              <button
                type="button"
                onClick={() => {
                  if (!newFact.trim()) return;
                  add.mutate(newFact.trim());
                  setNewFact("");
                }}
                disabled={!newFact.trim()}
                className="h-11 rounded-[9px] border border-line bg-surface px-3.5 text-[13px] disabled:opacity-50 md:h-9"
              >
                {t("memory.add")}
              </button>
            </div>

            <ul className="m-0 flex list-none flex-col rounded-xl border border-line bg-surface p-0">
              {facts?.map((fact, index) => (
                <li
                  key={fact.id}
                  className={[
                    "flex items-center gap-3.5 px-4 py-3.5",
                    index > 0 ? "border-t border-line" : "",
                  ].join(" ")}
                >
                  <span className="flex-1 text-[14.5px]">{fact.text}</span>
                  <span className="shrink-0 text-xs text-muted">
                    {formatRelative(fact.createdAt, locale)}
                  </span>
                  <button
                    type="button"
                    onClick={() => remove.mutate(fact.id)}
                    aria-label={t("memory.delete")}
                    className="flex size-11 shrink-0 items-center justify-center rounded-lg text-muted hover:bg-line/50 md:size-8"
                  >
                    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                      <path d="M4 7h16M9 7V5h6v2M6 7v13h12V7" />
                    </svg>
                  </button>
                </li>
              ))}
            </ul>
          </section>
        </div>
      </div>
    </>
  );
}
