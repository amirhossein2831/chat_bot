import type { ReactNode } from "react";
import { Navigate } from "react-router-dom";

import { useMe } from "./useAuth";

export function RequireAuth({ children }: { children: ReactNode }) {
  const { data, isLoading, isError } = useMe();

  if (isLoading) {
    return <div className="flex h-[100dvh] items-center justify-center text-muted">…</div>;
  }

  if (isError || !data) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}
