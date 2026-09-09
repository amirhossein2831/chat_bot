import { createBrowserRouter, RouterProvider } from "react-router-dom";

import { AppLayout } from "./AppLayout";

const router = createBrowserRouter([
  {
    path: "/",
    element: <AppLayout />,
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
