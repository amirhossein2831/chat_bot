export class ApiError extends Error {
  constructor(
    readonly code: string,
    readonly status: number,
    message: string,
    readonly fields?: FieldError[],
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export type FieldError = { field: string; rule: string };

export const isUnauthorized = (error: unknown): boolean =>
  error instanceof ApiError && error.code === "unauthorized";

export const isQuotaExceeded = (error: unknown): boolean =>
  error instanceof ApiError && error.code === "quota_exceeded";

export const isModelUnavailable = (error: unknown): boolean =>
  error instanceof ApiError && error.code === "model_unavailable";

export const isStorageUnavailable = (error: unknown): boolean =>
  error instanceof ApiError && error.code === "storage_unavailable";
