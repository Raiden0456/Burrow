import { isValidElement, type ReactNode } from "react"
import ReactMarkdown from "react-markdown"
import remarkGfm from "remark-gfm"
import { cn } from "@/lib/utils"
import { CodeBlock } from "@/components/CodeBlock"

type CodeProps = { className?: string; children?: ReactNode }

function Pre({ children }: { children?: ReactNode }) {
  if (!isValidElement<CodeProps>(children)) return <pre>{children}</pre>
  const lang = /language-([\w-]+)/.exec(children.props.className ?? "")?.[1] ?? "text"
  const code = String(children.props.children ?? "").replace(/\n$/, "")
  return <CodeBlock code={code} lang={lang} />
}

export function Markdown({ children, className }: { children: string; className?: string }) {
  return (
    <div className={cn("prose prose-theme max-w-none min-w-0 [overflow-wrap:anywhere] prose-pre:p-0 prose-code:rounded prose-code:bg-muted prose-code:px-1 prose-code:py-0.5 prose-code:font-normal prose-code:before:content-none prose-code:after:content-none", className)}>
      <ReactMarkdown remarkPlugins={[remarkGfm]} components={{ pre: Pre }}>
        {children}
      </ReactMarkdown>
    </div>
  )
}
