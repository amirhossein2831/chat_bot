import { memo } from "react";
import ReactMarkdown from "react-markdown";
import rehypeHighlight from "rehype-highlight";
import remarkGfm from "remark-gfm";

import type { PluggableList } from "unified";

import { CodeBlock } from "./CodeBlock";
import { Diagram } from "./Diagram";
import { useMathPlugins } from "./useMathPlugins";

// A half-arrived code fence renders as broken output that flickers when it
// completes. Counting fences and closing an odd one costs nothing.
type HastNode = { type?: string; value?: string; properties?: { className?: unknown }; children?: HastNode[] };

function textOf(node: unknown): string {
  const current = node as HastNode | undefined;
  if (!current) return "";
  if (current.type === "text") return current.value ?? "";
  return (current.children ?? []).map(textOf).join("");
}

function languageOf(node: unknown): string {
  const code = (node as HastNode | undefined)?.children?.[0];
  const names = code?.properties?.className;
  if (!Array.isArray(names)) return "text";

  const tag = names.find(
    (name): name is string => typeof name === "string" && name.startsWith("language-"),
  );
  return tag ? tag.slice("language-".length) : "text";
}

function balanceFences(markdown: string): string {
  const fences = markdown.match(/```/g)?.length ?? 0;
  return fences % 2 === 1 ? `${markdown}\n\`\`\`` : markdown;
}

// Memoised on the text alone. Rendering markdown replaces the whole subtree,
// so without this every refetch rebuilds the DOM of every answer in the
// thread, not just the one that changed.
export const Markdown = memo(function Markdown({ children }: { children: string }) {
  const math = useMathPlugins(children);

  const highlight: PluggableList[number] = [rehypeHighlight, { detect: true, ignoreMissing: true }];

  const remarkPlugins: PluggableList = math ? [remarkGfm, ...math.remark] : [remarkGfm];
  const rehypePlugins: PluggableList = math ? [...math.rehype, highlight] : [highlight];

  return (
    <div className="prose-chat flex flex-col gap-4 leading-[1.75]">
      <ReactMarkdown
        remarkPlugins={remarkPlugins}
        rehypePlugins={rehypePlugins}
        components={{
          // Code is always LTR, whatever direction surrounds it. The source is
          // read off the original node because by this point the rendered
          // children are highlighting spans, not the text someone wants copied.
          pre: ({ node, children }) => {
            const language = languageOf(node);
            const source = textOf(node);

            if (language === "mermaid") return <Diagram source={source} />;

            return (
              <CodeBlock language={language} source={source}>
                {children}
              </CodeBlock>
            );
          },
          code: ({ children }) => <code dir="ltr">{children}</code>,
          a: ({ href, children }) => (
            <a href={href} target="_blank" rel="noreferrer" className="text-accent underline">
              {children}
            </a>
          ),
          p: ({ children }) => <p className="m-0">{children}</p>,
        }}
      >
        {balanceFences(children)}
      </ReactMarkdown>
    </div>
  );
});
