import { useEffect, useState, type MouseEvent } from "react"
import { CheckIcon, CopyIcon } from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { useI18n } from "@/i18n"
import { copyText } from "@/lib/clipboard"
import { highlight } from "@/lib/highlight"

export function CodeBlock({ code, lang }: { code: string; lang: string }) {
  const [html, setHtml] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const { t } = useI18n()

  useEffect(() => {
    let alive = true
    highlight(code, lang).then((out) => alive && setHtml(out))
    return () => {
      alive = false
    }
  }, [code, lang])

  const copy = async (e: MouseEvent<HTMLButtonElement>) => {
    try {
      await copyText(code, e.currentTarget.parentElement ?? undefined)
    } catch {
      toast.error(t("common.copyFailed"))
      return
    }
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }

  return (
    <div className="code-block group relative my-4 overflow-hidden rounded-lg border text-sm not-prose">
      {html ? (
        <div dangerouslySetInnerHTML={{ __html: html }} />
      ) : (
        <pre className="shiki">
          <code>{code}</code>
        </pre>
      )}
      <Button
        size="icon-sm"
        variant="ghost"
        className="absolute top-2 right-2 opacity-0 transition-opacity group-hover:opacity-100 focus-visible:opacity-100 pointer-coarse:opacity-100"
        onClick={copy}
        aria-label={t("common.copy")}
      >
        {copied ? <CheckIcon /> : <CopyIcon />}
      </Button>
    </div>
  )
}
