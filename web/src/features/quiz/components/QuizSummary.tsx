import { Link } from "react-router"
import { CheckCircle2Icon, CircleAlertIcon, TrophyIcon, XCircleIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { useI18n } from "@/i18n"
import { cn } from "@/lib/utils"
import type { Quiz } from "../types/quiz"

export function QuizSummary({ quiz, onReview }: { quiz: Quiz; onReview: (index: number) => void }) {
  const { t } = useI18n()
  const answered = quiz.questions.filter((q) => q.result)
  const score = Math.round((answered.reduce((sum, q) => sum + (q.result?.score ?? 0), 0) / Math.max(answered.length, 1)) * 100)
  const correct = answered.filter((q) => q.result?.correct).length

  return (
    <Card>
      <CardContent className="flex flex-col gap-6">
        <div className="flex items-center gap-4 rounded-xl bg-gradient-to-br from-hue-yellow/20 via-hue-orange/10 to-transparent p-5">
          <span className="flex size-14 items-center justify-center rounded-2xl bg-hue-yellow/20 text-hue-yellow">
            <TrophyIcon className="size-7" />
          </span>
          <div>
            <div className="text-sm text-muted-foreground">{t("quiz.summary.title")}</div>
            <div className="flex items-baseline gap-3">
              <span className="text-4xl font-semibold tabular-nums">{score}%</span>
              <span className="text-muted-foreground">{t("quiz.summary.correct", { correct, total: quiz.questions.length })}</span>
            </div>
          </div>
        </div>
        <div data-glass className="flex flex-col divide-y overflow-hidden rounded-lg border">
          {quiz.questions.map((q, i) => {
            const r = q.result
            const Icon = r?.correct ? CheckCircle2Icon : r && r.score > 0 ? CircleAlertIcon : XCircleIcon
            const color = r?.correct ? "text-success" : r && r.score > 0 ? "text-warning" : "text-destructive"
            return (
              <button key={q.id} type="button" onClick={() => onReview(i)} className="flex items-start gap-3 px-3 py-2.5 text-left text-sm hover:bg-muted/60 pointer-coarse:py-3.5">
                <Icon className={cn("mt-0.5 size-4 shrink-0", color)} />
                <span className="line-clamp-2">{q.prompt.replace(/[`*#>]/g, "")}</span>
              </button>
            )
          })}
        </div>
        <div className="flex flex-wrap gap-2 max-md:flex-col">
          {quiz.topicSlug && (
            <Button asChild variant="outline" size="lg">
              <Link to={`/topics/${quiz.topicSlug}`}>{t("common.backToTopic")}</Link>
            </Button>
          )}
          <Button asChild size="lg">
            <Link to="/">{t("quiz.summary.dashboard")}</Link>
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
