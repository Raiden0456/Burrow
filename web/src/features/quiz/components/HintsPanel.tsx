import { LightbulbIcon, Loader2Icon, MessageSquareIcon, SparklesIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { ErrorAlert } from "@/components/ErrorAlert"
import { Markdown } from "@/components/Markdown"
import { useStartChat } from "@/features/chat"
import { useI18n } from "@/i18n"
import { cn } from "@/lib/utils"
import { useHintStream } from "../hooks/useHintStream"
import type { Question } from "../types/quiz"
import { questionContext } from "../utils/chatContext"

const maxHints = 3

function HintCard({ kind, index, children }: { kind: "hint" | "explain"; index: number; children: string }) {
  const { t } = useI18n()
  const hint = kind === "hint"
  return (
    <div className={cn("rounded-lg border p-3", hint ? "border-hue-yellow/40 bg-hue-yellow/8" : "border-hue-purple/40 bg-hue-purple/8")}>
      <div className={cn("mb-1 flex items-center gap-1.5 text-xs font-medium tracking-wide uppercase", hint ? "text-hue-yellow" : "text-hue-purple")}>
        {hint ? <LightbulbIcon className="size-3.5" /> : <SparklesIcon className="size-3.5" />}
        {hint ? t("quiz.hintN", { n: index }) : t("quiz.deepDive")}
      </div>
      <Markdown className="prose-sm prose-headings:text-sm prose-headings:font-semibold">{children}</Markdown>
    </div>
  )
}

export function HintsPanel({ quizId, question }: { quizId: number; question: Question }) {
  const { t } = useI18n()
  const stream = useHintStream(quizId, question.id)
  const startChat = useStartChat()
  const answered = !!question.result
  const hints = question.hints.filter((h) => h.kind === "hint")
  const explained = question.hints.some((h) => h.kind === "explain")
  const busy = stream.streaming !== null

  return (
    <div className="flex flex-col gap-3">
      {question.hints.map((h) => (
        <HintCard key={h.id} kind={h.kind} index={hints.indexOf(h) + 1}>
          {h.content}
        </HintCard>
      ))}
      {busy && (
        <HintCard kind={stream.streaming ?? "hint"} index={hints.length + 1}>
          {stream.text || "…"}
        </HintCard>
      )}
      {stream.error && <ErrorAlert error={stream.error} />}

      <div className="flex flex-wrap gap-2 max-md:flex-col">
        {!answered && hints.length < maxHints && (
          <Button variant="outline" size="sm" disabled={busy} onClick={() => stream.start("hint")} className="border-hue-yellow/40 text-hue-yellow hover:bg-hue-yellow/10 hover:text-hue-yellow">
            {busy ? <Loader2Icon className="animate-spin" /> : <LightbulbIcon />}
            {hints.length === 0 ? t("quiz.hint") : t("quiz.moreHint", { n: hints.length + 1, max: maxHints })}
          </Button>
        )}
        {answered && !explained && (
          <Button variant="outline" size="sm" disabled={busy} onClick={() => stream.start("explain")} className="border-hue-purple/40 text-hue-purple hover:bg-hue-purple/10 hover:text-hue-purple">
            {busy ? <Loader2Icon className="animate-spin" /> : <SparklesIcon />}
            {t("quiz.explainMore")}
          </Button>
        )}
        {answered && (
          <Button
            variant="ghost"
            size="sm"
            disabled={startChat.isPending}
            onClick={() => startChat.mutate({ title: question.prompt.slice(0, 60), context: questionContext(question) })}
          >
            <MessageSquareIcon />
            {t("quiz.discuss")}
          </Button>
        )}
      </div>
    </div>
  )
}
