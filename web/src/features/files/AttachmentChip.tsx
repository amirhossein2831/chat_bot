import { useTranslation } from "react-i18next";

import type { UploadItem } from "./useUpload";

type Props = {
  item: UploadItem;
  onRemove: (id: string) => void;
};

export function AttachmentChip({ item, onRemove }: Props) {
  const { t } = useTranslation();

  return (
    <div className="flex items-center gap-2 rounded-[9px] border border-line bg-surface px-2.5 py-1.5 text-[12.5px]">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="shrink-0 text-muted">
        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
        <path d="M14 2v6h6" />
      </svg>

      <span className="max-w-[160px] truncate">{item.filename}</span>
      <span className={statusClass(item)}>{statusLabel(item, t)}</span>

      <button
        type="button"
        onClick={() => onRemove(item.id)}
        aria-label={t("memory.delete")}
        className="flex size-6 shrink-0 items-center justify-center rounded text-muted hover:bg-line/60"
      >
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.4" strokeLinecap="round">
          <path d="M6 6l12 12M18 6 6 18" />
        </svg>
      </button>
    </div>
  );
}

function statusLabel(item: UploadItem, t: (key: string) => string): string {
  if (item.uploading || item.status === "pending" || item.status === "queued") {
    return t("files.extracting");
  }
  if (item.status === "done") return t("files.ready");
  if (item.status === "unsupported") return t("files.unsupported");
  return t("files.failed");
}

function statusClass(item: UploadItem): string {
  if (item.status === "done") return "shrink-0 text-ok";
  if (item.status === "failed") return "shrink-0 text-red-600 dark:text-red-400";
  return "shrink-0 text-muted";
}
