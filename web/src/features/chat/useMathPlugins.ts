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
  ]).then(([remark, rehype]) => {
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
