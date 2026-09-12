import type { Locale } from "./direction";

// The platform knows the Persian calendar, so no date library is needed.
// Formatting is display-only and never round-trips into storage.
export function formatDate(iso: string, locale: Locale): string {
  const date = new Date(iso);

  if (locale === "fa") {
    return new Intl.DateTimeFormat("fa-IR-u-ca-persian", {
      day: "numeric",
      month: "long",
      hour: "2-digit",
      minute: "2-digit",
    }).format(date);
  }

  return new Intl.DateTimeFormat("en-GB", {
    day: "numeric",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

export function formatRelative(iso: string, locale: Locale): string {
  const elapsed = Date.now() - new Date(iso).getTime();
  const formatter = new Intl.RelativeTimeFormat(locale === "fa" ? "fa-IR" : "en-GB", {
    numeric: "auto",
  });

  const minutes = Math.round(elapsed / 60_000);
  if (minutes < 60) return formatter.format(-minutes, "minute");

  const hours = Math.round(minutes / 60);
  if (hours < 24) return formatter.format(-hours, "hour");

  return formatter.format(-Math.round(hours / 24), "day");
}

export function formatNumber(value: number, locale: Locale): string {
  return new Intl.NumberFormat(locale === "fa" ? "fa-IR" : "en-GB").format(value);
}
