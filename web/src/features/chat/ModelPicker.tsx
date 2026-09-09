import { useState } from "react";
import { useTranslation } from "react-i18next";

import { useModels } from "./useModels";

type Props = {
  value: string;
  onChange: (modelId: string) => void;
};

export function ModelPicker({ value, onChange }: Props) {
  const { t } = useTranslation();
  const { data: models } = useModels();
  const [open, setOpen] = useState(false);

  const selected = models?.find((model) => model.id === value) ?? models?.find((m) => m.available);

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen((current) => !current)}
        className="flex h-11 items-center gap-2 rounded-[9px] border border-line bg-surface px-3 text-[13px] md:h-9"
      >
        <span
          className={[
            "size-1.5 rounded-full",
            selected?.available ? "bg-ok" : "bg-muted",
          ].join(" ")}
        />
        <span className="font-medium">{selected?.id ?? "…"}</span>
        {selected?.tier === "local" && (
          <span className="text-muted">· {t("models.local")}</span>
        )}
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-muted">
          <path d="m6 9 6 6 6-6" />
        </svg>
      </button>

      {open && (
        <>
          <button
            type="button"
            aria-hidden
            tabIndex={-1}
            onClick={() => setOpen(false)}
            className="fixed inset-0 z-10"
          />
          <ul className="absolute z-20 mt-1.5 flex w-[280px] list-none flex-col gap-0.5 rounded-xl border border-line bg-surface p-1.5 shadow-lg end-0">
            {models?.map((model) => (
              <li key={model.id}>
                <button
                  type="button"
                  // An unavailable model stays visible and disabled: a user
                  // should see that it exists and is down, not wonder where it
                  // went.
                  disabled={!model.available}
                  onClick={() => {
                    onChange(model.id);
                    setOpen(false);
                  }}
                  className="flex w-full flex-col gap-0.5 rounded-lg px-3 py-2.5 text-start hover:bg-raised disabled:opacity-50 disabled:hover:bg-transparent"
                >
                  <span className="flex items-center gap-2 text-[13.5px] font-medium">
                    <span
                      className={[
                        "size-1.5 rounded-full",
                        model.available ? "bg-ok" : "bg-muted",
                      ].join(" ")}
                    />
                    {model.id}
                  </span>
                  <span className="text-xs text-muted">
                    {model.available
                      ? model.tier === "local"
                        ? t("models.local")
                        : model.provider
                      : t("models.unavailable")}
                  </span>
                </button>
              </li>
            ))}
          </ul>
        </>
      )}
    </div>
  );
}
