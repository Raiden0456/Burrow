import { useI18n } from "@/i18n"
import { cn } from "@/lib/utils"
import type { Question } from "../types/quiz"

type Props = {
  questions: Question[]
  current: number
  onSelect: (i: number) => void
  className?: string
}

export function QuizStepper({ questions, current, onSelect, className }: Props) {
  const { t } = useI18n()
  return (
    <div className={cn("flex flex-wrap gap-1.5 max-md:flex-nowrap max-md:gap-1", className)}>
      {questions.map((q, i) => (
        <button
          key={q.id}
          type="button"
          aria-label={t("quiz.questionAria", { n: i + 1 })}
          onClick={() => onSelect(i)}
          className="flex h-6 items-center max-md:h-11 max-md:min-w-0 max-md:flex-1 md:h-auto"
        >
          <span
            className={cn(
              "block h-2 w-8 rounded-full bg-muted transition-colors max-md:w-full",
              q.result && (q.result.correct ? "bg-success-solid" : q.result.score > 0 ? "bg-warning-solid" : "bg-destructive"),
              i === current && "ring-2 ring-ring ring-offset-2 ring-offset-background",
            )}
          />
        </button>
      ))}
    </div>
  )
}
