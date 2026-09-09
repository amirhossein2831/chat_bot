import { memo } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

// A half-arrived code fence renders as broken output that flickers when it
// completes. Counting fences and closing an odd one costs nothing.
function balanceFences(markdown: string): string {
  const fences = markdown.match(/```/g)?.length ?? 0;
  return fences % 2 === 1 ? `${markdown}\n\`\`\`` : markdown;
}

// Memoised on the text alone. Rendering markdown replaces the whole subtree,
// so without this every refetch rebuilds the DOM of every answer in the
// thread, not just the one that changed.
export const Markdown = memo(function Markdown({ children }: { children: string }) {
  return (
    <div className="prose-chat flex flex-col gap-4 leading-[1.75]">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          // Code is always LTR, whatever direction surrounds it.
          pre: ({ children }) => (
            <pre
              dir="ltr"
              className="overflow-x-auto rounded-[14px] bg-raised p-4 font-mono text-[13px] leading-relaxed"
            >
              {children}
            </pre>
          ),
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
