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
    <div className="flex h-[100dvh] overflow-hidden bg-surface">
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
          "fixed inset-y-0 z-30 w-[286px] shrink-0 bg-surface",
          "transition-transform duration-200 lg:static lg:translate-x-0",
          open ? "translate-x-0" : "translate-x-full lg:translate-x-0",
        ].join(" ")}
      >
        <Sidebar />
      </aside>

      {/* min-h-0 is load-bearing: without it this column refuses to shrink
          below its content, the message list never becomes scrollable, and the
          thread grows behind the composer instead. */}
      <main className="flex min-h-0 min-w-0 flex-1 flex-col bg-paper lg:my-2 lg:me-2 lg:rounded-[18px] lg:shadow-[var(--shadow-soft)]">
        <Outlet />
      </main>
    </div>
  );
}
