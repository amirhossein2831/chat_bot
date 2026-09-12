import { lazy, Suspense } from "react";
import { createBrowserRouter, RouterProvider } from "react-router-dom";

import { LoginPage } from "../features/auth/LoginPage";
import { RequireAuth } from "../features/auth/RequireAuth";
import { AppLayout } from "./AppLayout";

// Chat carries the markdown renderer and highlighting; prompts and memory are
// visited rarely. Splitting them keeps the first paint small, which is what a
// user on a mobile connection actually feels.
const ChatPage = lazy(() =>
  import("../features/chat/ChatPage").then((m) => ({ default: m.ChatPage })),
);
const PromptsPage = lazy(() =>
  import("../features/prompts/PromptsPage").then((m) => ({ default: m.PromptsPage })),
);
const MemoryPage = lazy(() =>
  import("../features/memory/MemoryPage").then((m) => ({ default: m.MemoryPage })),
);

function Loading() {
  return <div className="flex flex-1 items-center justify-center text-muted">…</div>;
}

function lazyRoute(element: React.ReactNode) {
  return <Suspense fallback={<Loading />}>{element}</Suspense>;
}

const router = createBrowserRouter([
  { path: "/login", element: <LoginPage /> },
  {
    path: "/",
    element: (
      <RequireAuth>
        <AppLayout />
      </RequireAuth>
    ),
    children: [
      { index: true, element: lazyRoute(<ChatPage />) },
      { path: "c/:id", element: lazyRoute(<ChatPage />) },
      { path: "prompts", element: lazyRoute(<PromptsPage />) },
      { path: "memory", element: lazyRoute(<MemoryPage />) },
    ],
  },
]);

export function Router() {
  return <RouterProvider router={router} />;
}
