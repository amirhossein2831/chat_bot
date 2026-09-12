import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Navigate } from "react-router-dom";

import { ApiError } from "../../lib/errors";
import { useLogin, useMe, useRegister } from "./useAuth";

type Mode = "login" | "register";

export function LoginPage() {
  const { t } = useTranslation();
  const { data: me } = useMe();

  const [mode, setMode] = useState<Mode>("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [invite, setInvite] = useState("");

  const loginMutation = useLogin();
  const registerMutation = useRegister();
  const active = mode === "login" ? loginMutation : registerMutation;

  if (me) return <Navigate to="/" replace />;

  const submit = (event: React.FormEvent) => {
    event.preventDefault();
    if (mode === "login") {
      loginMutation.mutate({ email, password });
    } else {
      registerMutation.mutate({ email, password, invite });
    }
  };

  return (
    <div className="flex min-h-[100dvh] items-center justify-center px-5 py-10">
      <div className="w-full max-w-[380px]">
        <div className="mb-8 flex items-center gap-2.5">
          <div className="flex size-8 items-center justify-center rounded-lg bg-accent">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.1" strokeLinecap="round" strokeLinejoin="round">
              <path d="M21 11.5a8.4 8.4 0 0 1-9 8.4 8.9 8.9 0 0 1-3.8-.9L3 21l1.9-5.1A8.4 8.4 0 0 1 12 3a8.4 8.4 0 0 1 9 8.5z" />
            </svg>
          </div>
          <span className="text-lg font-semibold tracking-tight">{t("app.name")}</span>
        </div>

        <form onSubmit={submit} className="flex flex-col gap-4">
          <Field
            label={t("auth.email")}
            value={email}
            onChange={setEmail}
            type="email"
            autoComplete="email"
            error={fieldError(active.error, "email")}
          />

          <Field
            label={t("auth.password")}
            value={password}
            onChange={setPassword}
            type="password"
            autoComplete={mode === "login" ? "current-password" : "new-password"}
            error={fieldError(active.error, "password")}
          />

          {mode === "register" && (
            <Field
              label={t("auth.inviteOptional")}
              hint={t("auth.inviteHelp")}
              value={invite}
              onChange={setInvite}
              error={fieldError(active.error, "invite")}
            />
          )}

          {generalError(active.error, mode) && (
            <p className="text-sm text-red-600 dark:text-red-400">
              {generalError(active.error, mode)}
            </p>
          )}

          <button
            type="submit"
            disabled={active.isPending}
            className="mt-1 h-11 rounded-lg bg-accent text-[15px] font-medium text-white disabled:opacity-60"
          >
            {mode === "login" ? t("auth.login") : t("auth.register")}
          </button>
        </form>

        <button
          type="button"
          onClick={() => setMode(mode === "login" ? "register" : "login")}
          className="mt-5 text-sm text-muted hover:text-ink"
        >
          {mode === "login" ? t("auth.noAccount") : t("auth.haveAccount")}
        </button>
      </div>
    </div>
  );
}

type FieldProps = {
  label: string;
  value: string;
  onChange: (value: string) => void;
  type?: string;
  autoComplete?: string;
  error?: string;
  hint?: string;
};

function Field({ label, value, onChange, type = "text", autoComplete, error, hint }: FieldProps) {
  return (
    <label className="flex flex-col gap-1.5">
      <span className="text-[13px] font-medium text-muted">{label}</span>
      <input
        type={type}
        value={value}
        autoComplete={autoComplete}
        onChange={(event) => onChange(event.target.value)}
        className="h-11 rounded-lg border border-line bg-surface px-3 text-[15px] outline-none focus:border-accent"
      />
      {hint && !error && <span className="text-xs text-muted">{hint}</span>}
      {error && <span className="text-xs text-red-600 dark:text-red-400">{error}</span>}
    </label>
  );
}

// The gateway returns a field and a rule per invalid input rather than a
// sentence, so the message is ours to write and translate.
function fieldError(error: unknown, field: string): string | undefined {
  if (!(error instanceof ApiError) || !error.fields) return undefined;
  return error.fields.find((f) => f.field === field) ? ruleMessage(field) : undefined;
}

function ruleMessage(field: string): string {
  if (field === "email") return "ایمیل معتبر نیست";
  if (field === "password") return "گذرواژه باید حداقل ۱۰ نویسه باشد";
  return "این فیلد لازم است";
}

// The gateway returns unauthorized for two different things: a wrong password
// on login, and an invalid or already-used invite on registration. Only the
// caller knows which one it asked for.
function generalError(error: unknown, mode: Mode): string | undefined {
  if (!(error instanceof ApiError)) return undefined;
  if (error.fields?.length) return undefined;

  if (error.code === "unauthorized") {
    return mode === "register"
      ? "این کد دعوت نامعتبر است یا قبلاً استفاده شده. می‌توانید بدون کد ثبت‌نام کنید."
      : "ایمیل یا گذرواژه درست نیست";
  }
  if (error.code === "conflict") return "این ایمیل قبلاً ثبت شده است";
  return "خطایی رخ داد";
}
