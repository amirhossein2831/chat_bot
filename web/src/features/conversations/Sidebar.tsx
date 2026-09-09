import { useTranslation } from "react-i18next";

export function Sidebar() {
  const { t } = useTranslation();

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2.5 px-[18px] pt-5">
        <div className="flex size-[26px] items-center justify-center rounded-[7px] bg-accent">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.1" strokeLinecap="round" strokeLinejoin="round">
            <path d="M21 11.5a8.4 8.4 0 0 1-9 8.4 8.9 8.9 0 0 1-3.8-.9L3 21l1.9-5.1A8.4 8.4 0 0 1 12 3a8.4 8.4 0 0 1 9 8.5z" />
          </svg>
        </div>
        <span className="text-[15.5px] font-semibold tracking-tight">{t("app.name")}</span>
      </div>
    </div>
  );
}
