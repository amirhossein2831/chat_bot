import { useEffect, useState } from "react";
import type { PluggableList } from "unified";

type Plugins = { remark: PluggableList; rehype: PluggableList };

// Loaded once for the session and shared: KaTeX and its stylesheet are a
// sizeable download, and most answers contain no formula at all.
let loaded: Plugins | null = null;
let loading: Promise<Plugins> | null = null;

function load(): Promise<Plugins> {
  loading ??= Promise.all([
    import("remark-math"),
    import("rehype-katex"),
    import("katex/dist/katex.min.css"),
  ])
    .then(async ([remark, rehype]) => {
      await warmFonts();
      return [remark, rehype] as const;
    })
    .then(([remark, rehype]) => {
    loaded = {
      remark: [remark.default],
      // trust is off and errors are not thrown: formulas come from a model, so
      // they are untrusted input and are often half-written.
      rehype: [[rehype.default, { trust: false, throwOnError: false, errorColor: "currentColor" }]],
    };
    return loaded;
  });
  return loading;
}

// The radical and the large delimiters live in the Size fonts. A formula
// painted before they arrive is drawn from the fallback serif at the wrong
// size, which is what makes a square root sit on top of what it covers.
const FAMILIES = [
  "KaTeX_Main",
  "KaTeX_Math",
  "KaTeX_Size1",
  "KaTeX_Size2",
  "KaTeX_Size3",
  "KaTeX_Size4",
];

async function warmFonts(): Promise<void> {
  if (typeof document === "undefined" || !document.fonts) return;

  // Asking for a face the stylesheet has not registered yet resolves without
  // loading anything, and the import resolves before the rules are applied.
  for (let attempt = 0; attempt < 20; attempt += 1) {
    if ([...document.fonts].some((face) => face.family === "KaTeX_Size1")) break;
    await new Promise((resolve) => setTimeout(resolve, 25));
  }

  await Promise.all(FAMILIES.map((family) => document.fonts.load(`16px ${family}`))).catch(
    () => undefined,
  );
}

export function useMathPlugins(text: string): Plugins | null {
  const wanted = text.includes("$");
  const [plugins, setPlugins] = useState(loaded);

  useEffect(() => {
    if (!wanted || loaded) return;

    let cancelled = false;
    void load().then((ready) => {
      if (!cancelled) setPlugins(ready);
    });
    return () => {
      cancelled = true;
    };
  }, [wanted]);

  return wanted ? plugins : null;
}
