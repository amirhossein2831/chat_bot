import { Outlet } from "react-router-dom";

import { Sidebar } from "../features/conversations/Sidebar";
import { useUiStore } from "./store";

// One layout at three widths. The sidebar is an overlay below lg and a
// persistent column above it. No component branches on a device: two
// implementations diverge within a month.
export function AppLayout() {
  const open = useUiStore((state) => state.sidebarOpen);
  const setOpen = useUiStore((state) => state.setSidebarOpen);

  return (
    <div className="flex h-[100dvh] overflow-hidden">
      {open && (
        <button
          type="button"
          aria-hidden
          tabIndex={-1}
          onClick={() => setOpen(false)}
          className="fixed inset-0 z-20 bg-ink/20 lg:hidden"
        />
      )}

      <aside
        className={[
          "fixed inset-y-0 z-30 w-[296px] shrink-0 border-s border-line bg-raised",
          "transition-transform duration-200 lg:static lg:translate-x-0",
          open ? "translate-x-0" : "translate-x-full lg:translate-x-0",
        ].join(" ")}
      >
        <Sidebar />
      </aside>

      <main className="flex min-w-0 flex-1 flex-col">
        <Outlet />
      </main>
    </div>
  );
}
