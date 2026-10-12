import { useI18n } from "@/i18n"
import { cn } from "@/lib/utils"
import type { Question } from "../types/quiz"

export function QuizStepper({ questions, current, onSelect }: { questions: Question[]; current: number; onSelect: (i: number) => void }) {
  const { t } = useI18n()
  return (
    <div className="flex flex-wrap gap-1.5">
      {questions.map((q, i) => (
        <button
          key={q.id}
          type="button"
          aria-label={t("quiz.questionAria", { n: i + 1 })}
          onClick={() => onSelect(i)}
          className={cn(
            "h-2 w-8 rounded-full bg-muted transition-colors",
            q.result && (q.result.correct ? "bg-success" : q.result.score > 0 ? "bg-warning" : "bg-destructive"),
            i === current && "ring-2 ring-ring ring-offset-2 ring-offset-background",
          )}
        />
      ))}
    </div>
  )
}
