import { CheckCircle2Icon, CircleAlertIcon, LightbulbIcon, XCircleIcon, BookCheckIcon } from "lucide-react"
import { Markdown } from "@/components/Markdown"
import { useI18n } from "@/i18n"
import { cn } from "@/lib/utils"
import type { Question } from "../types/quiz"

export function AnswerFeedback({ question }: { question: Question }) {
  const { t } = useI18n()
  const r = question.result
  if (!r) return null

  const partial = !r.correct && r.score > 0
  const Icon = r.correct ? CheckCircle2Icon : partial ? CircleAlertIcon : XCircleIcon
  const verdict = r.correct ? t("quiz.verdict.correct") : partial ? t("quiz.verdict.partial") : t("quiz.verdict.wrong")

  return (
    <div className="flex flex-col gap-3">
      <div
        className={cn(
          "flex items-center gap-2 rounded-lg px-3 py-2 font-medium",
          r.correct ? "bg-success/12 text-success" : partial ? "bg-warning/12 text-warning" : "bg-destructive/12 text-destructive",
        )}
      >
        <Icon className="size-5" />
        {verdict}
        <span className="text-sm font-normal opacity-80">· {Math.round(r.score * 100)}%</span>
      </div>
      {r.feedback && <Markdown className="prose-sm">{r.feedback}</Markdown>}
      {question.kind === "open" && (
        <div className="rounded-lg border border-hue-green/30 bg-hue-green/5 p-3">
          <div className="mb-1 flex items-center gap-1.5 text-xs font-medium tracking-wide text-hue-green uppercase">
            <BookCheckIcon className="size-3.5" />
            {t("quiz.referenceAnswer")}
          </div>
          <Markdown className="prose-sm">{r.referenceAnswer}</Markdown>
        </div>
      )}
      <div className="rounded-lg border border-hue-blue/30 bg-hue-blue/5 p-3">
        <div className="mb-1 flex items-center gap-1.5 text-xs font-medium tracking-wide text-hue-blue uppercase">
          <LightbulbIcon className="size-3.5" />
          {t("quiz.explanation")}
        </div>
        <Markdown className="prose-sm">{r.explanation}</Markdown>
      </div>
    </div>
  )
}
