import { LifeBuoyIcon } from "lucide-react"
import { ErrorAlert } from "@/components/ErrorAlert"
import { Markdown } from "@/components/Markdown"
import { useI18n } from "@/i18n"
import type { DrillHint } from "../types/drill"

type Props = {
  hints: DrillHint[]
  streamingText: string | null
  error: string | null
}

export function HelpPanel({ hints, streamingText, error }: Props) {
  const { t } = useI18n()
  if (hints.length === 0 && streamingText === null && !error) return null

  const items = [...hints.map((h) => ({ key: String(h.id), content: h.content })), ...(streamingText !== null ? [{ key: "streaming", content: streamingText || "…" }] : [])]

  return (
    <div className="flex flex-col gap-3">
      {items
        .slice()
        .reverse()
        .map((item, i) => (
          <div key={item.key} data-glass className="rounded-lg border border-hue-yellow/40 bg-hue-yellow/8 p-3">
            <div className="mb-1 flex items-center gap-1.5 text-xs font-medium tracking-wide text-hue-yellow uppercase">
              <LifeBuoyIcon className="size-3.5" />
              {t("drills.helpN", { n: items.length - i })}
            </div>
            <Markdown className="prose-sm prose-headings:text-sm prose-headings:font-semibold">{item.content}</Markdown>
          </div>
        ))}
      {error && <ErrorAlert error={error} />}
    </div>
  )
}
