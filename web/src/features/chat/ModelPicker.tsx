import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";

import { ProviderIcon } from "./ProviderIcon";
import type { Model } from "./useModels";
import { useModels } from "./useModels";

type Props = {
  value: string;
  onChange: (modelId: string) => void;
};

type Group = {
  provider: string;
  models: Model[];
  available: number;
};

export function ModelPicker({ value, onChange }: Props) {
  const { t } = useTranslation();
  const { data: models } = useModels();

  const [open, setOpen] = useState(false);
  const [provider, setProvider] = useState<string | null>(null);

  const groups = useMemo(() => groupByProvider(models ?? []), [models]);
  const selected = models?.find((model) => model.id === value) ?? models?.find((m) => m.available);
  const opened = groups.find((group) => group.provider === provider);

  const close = () => {
    setOpen(false);
    setProvider(null);
  };

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => (open ? close() : setOpen(true))}
        className="flex h-11 items-center gap-2 rounded-[9px] border border-line bg-surface px-3 text-[13px] md:h-9"
      >
        {selected && <ProviderIcon provider={selected.provider} />}
        <Dot available={selected?.available} />
        <span className="font-medium">{selected?.id ?? "…"}</span>
        {selected?.tier === "local" && <span className="text-muted">· {t("models.local")}</span>}
        <Chevron className="text-muted" />
      </button>

      {open && (
        <>
          <button
            type="button"
            aria-hidden
            tabIndex={-1}
            onClick={close}
            className="fixed inset-0 z-10"
          />
          <div className="absolute z-20 mt-1.5 w-[280px] rounded-xl border border-line bg-surface p-1.5 shadow-lg end-0">
            {opened ? (
              <>
                <button
                  type="button"
                  onClick={() => setProvider(null)}
                  className="mb-1 flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-[13px] font-medium hover:bg-raised"
                >
                  <Chevron className="rotate-90 text-muted rtl:-rotate-90" />
                  <ProviderIcon provider={opened.provider} />
                  {t(`models.providers.${opened.provider}`, opened.provider)}
                </button>

                <ul className="flex list-none flex-col gap-0.5">
                  {opened.models.map((model) => (
                    <li key={model.id}>
                      <Row
                        // An unavailable model stays visible and disabled: a
                        // user should see that it exists and is down, not
                        // wonder where it went.
                        disabled={!model.available}
                        available={model.available}
                        title={model.id}
                        subtitle={
                          model.available
                            ? `${model.contextWindow.toLocaleString()} tokens`
                            : t("models.unavailable")
                        }
                        onClick={() => {
                          onChange(model.id);
                          close();
                        }}
                      />
                    </li>
                  ))}
                </ul>
              </>
            ) : (
              <ul className="flex list-none flex-col gap-0.5">
                {groups.map((group) => (
                  <li key={group.provider}>
                    <Row
                      available={group.available > 0}
                      leading={
                        <ProviderIcon
                          provider={group.provider}
                          className={[
                            "size-[18px] shrink-0",
                            group.available > 0 ? "" : "text-muted",
                          ].join(" ")}
                        />
                      }
                      title={t(`models.providers.${group.provider}`, group.provider)}
                      subtitle={
                        group.available > 0
                          ? t("models.countAvailable", { count: group.available })
                          : t("models.noneAvailable")
                      }
                      onClick={() => setProvider(group.provider)}
                      trailing={<Chevron className="-rotate-90 text-muted rtl:rotate-90" />}
                    />
                  </li>
                ))}
              </ul>
            )}
          </div>
        </>
      )}
    </div>
  );
}

// A provider with nothing reachable sinks to the bottom, so the list opens on
// something the user can actually pick.
function groupByProvider(models: Model[]): Group[] {
  const byProvider = new Map<string, Model[]>();
  for (const model of models) {
    byProvider.set(model.provider, [...(byProvider.get(model.provider) ?? []), model]);
  }

  return [...byProvider.entries()]
    .map(([provider, list]) => ({
      provider,
      models: list,
      available: list.filter((model) => model.available).length,
    }))
    .sort((a, b) => Number(b.available > 0) - Number(a.available > 0));
}

type RowProps = {
  title: string;
  subtitle: string;
  available: boolean;
  disabled?: boolean;
  onClick: () => void;
  leading?: React.ReactNode;
  trailing?: React.ReactNode;
};

function Row({ title, subtitle, available, disabled, onClick, leading, trailing }: RowProps) {
  return (
    <button
      type="button"
      disabled={disabled}
      onClick={onClick}
      className="flex w-full items-center gap-2 rounded-lg px-3 py-2.5 text-start hover:bg-raised disabled:opacity-50 disabled:hover:bg-transparent"
    >
      {leading ?? <Dot available={available} />}
      <span className="flex min-w-0 flex-1 flex-col gap-0.5">
        <span className="truncate text-[13.5px] font-medium">{title}</span>
        <span className="truncate text-xs text-muted">{subtitle}</span>
      </span>
      {trailing}
    </button>
  );
}

function Dot({ available }: { available?: boolean }) {
  return (
    <span
      className={["size-1.5 shrink-0 rounded-full", available ? "bg-ok" : "bg-muted"].join(" ")}
    />
  );
}

function Chevron({ className }: { className?: string }) {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
    >
      <path d="m6 9 6 6 6-6" />
    </svg>
  );
}
