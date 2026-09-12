export type Locale = "fa" | "en";

export function applyDirection(locale: Locale): void {
  const root = document.documentElement;
  root.lang = locale;
  root.dir = locale === "fa" ? "rtl" : "ltr";
}
