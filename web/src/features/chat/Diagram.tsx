import { useEffect, useState } from "react";

type Props = {
  source: string;
};

type State = { svg: string } | { failed: true } | null;

let counter = 0;

// Mermaid is several times the size of the whole application, so it is fetched
// only when an answer actually contains a diagram, and never before.
export function Diagram({ source }: Props) {
  const [state, setState] = useState<State>(null);

  useEffect(() => {
    let cancelled = false;

    const draw = async () => {
      try {
        const mermaid = (await import("mermaid")).default;

        mermaid.initialize({
          startOnLoad: false,
          // The diagram is written by a model: strict keeps it from emitting
          // markup or links of its own.
          securityLevel: "strict",
          theme: window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "default",
          fontFamily: "inherit",
        });

        counter += 1;
        const { svg } = await mermaid.render(`diagram-${counter}`, source);
        if (!cancelled) setState({ svg });
      } catch {
        // A half-arrived diagram cannot parse, and neither can a malformed
        // one. Both show as the code they are until they can be drawn.
        if (!cancelled) setState({ failed: true });
      }
    };

    void draw();
    return () => {
      cancelled = true;
    };
  }, [source]);

  if (state === null || "failed" in state) {
    return (
      <pre
        dir="ltr"
        className="overflow-x-auto rounded-[14px] bg-raised p-4 font-mono text-[13px] leading-relaxed"
      >
        {source}
      </pre>
    );
  }

  return (
    <div
      dir="ltr"
      className="flex justify-center overflow-x-auto rounded-[14px] bg-raised p-4"
      dangerouslySetInnerHTML={{ __html: state.svg }}
    />
  );
}
