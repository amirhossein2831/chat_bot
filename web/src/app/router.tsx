import { createBrowserRouter, RouterProvider } from "react-router-dom";

import { LoginPage } from "../features/auth/LoginPage";
import { ChatPage } from "../features/chat/ChatPage";
import { MemoryPage } from "../features/memory/MemoryPage";
import { PromptsPage } from "../features/prompts/PromptsPage";
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
    children: [
      { index: true, element: <Placeholder /> },
      { path: "c/:id", element: <ChatPage /> },
      { path: "prompts", element: <PromptsPage /> },
      { path: "memory", element: <MemoryPage /> },
    ],
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
