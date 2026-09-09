import { useEffect, useState } from "react";

// On a phone the keyboard shrinks the visual viewport without changing the
// layout viewport, so a bottom-anchored composer ends up behind it. No media
// query can see this; visualViewport can.
export function useVisualViewport(): number {
  const [inset, setInset] = useState(0);

  useEffect(() => {
    const viewport = window.visualViewport;
    if (!viewport) return;

    const update = () => {
      setInset(Math.max(0, window.innerHeight - viewport.height - viewport.offsetTop));
    };

    viewport.addEventListener("resize", update);
    viewport.addEventListener("scroll", update);
    update();

    return () => {
      viewport.removeEventListener("resize", update);
      viewport.removeEventListener("scroll", update);
    };
  }, []);

  return inset;
}
