import { createBrowserRouter, RouterProvider } from "react-router-dom";

import { LoginPage } from "../features/auth/LoginPage";
import { RequireAuth } from "../features/auth/RequireAuth";
import { AppLayout } from "./AppLayout";

const router = createBrowserRouter([
  { path: "/login", element: <LoginPage /> },
  {
    path: "/",
    element: (
      <RequireAuth>
        <AppLayout />
      </RequireAuth>
    ),
    children: [{ index: true, element: <Placeholder /> }],
  },
]);

function Placeholder() {
  return (
    <div className="flex flex-1 items-center justify-center text-muted">
      <span className="text-sm">…</span>
    </div>
  );
}

export function Router() {
  return <RouterProvider router={router} />;
}
