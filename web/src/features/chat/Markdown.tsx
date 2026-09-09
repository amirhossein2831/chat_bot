import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

// A half-arrived code fence renders as broken output that flickers when it
// completes. Counting fences and closing an odd one costs nothing.
function balanceFences(markdown: string): string {
  const fences = markdown.match(/```/g)?.length ?? 0;
  return fences % 2 === 1 ? `${markdown}\n\`\`\`` : markdown;
}

export function Markdown({ children }: { children: string }) {
  return (
    <div className="flex flex-col gap-3 text-[15px] [&_ol]:flex [&_ol]:list-decimal [&_ol]:flex-col [&_ol]:gap-2 [&_ol]:ps-6 [&_ul]:flex [&_ul]:list-disc [&_ul]:flex-col [&_ul]:gap-2 [&_ul]:ps-6">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          // Code is always LTR, whatever direction surrounds it.
          pre: ({ children }) => (
            <pre
              dir="ltr"
              className="overflow-x-auto rounded-[10px] border border-line bg-raised p-3.5 font-mono text-[12.5px] leading-relaxed"
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
}
