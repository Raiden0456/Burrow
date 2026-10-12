import { CheckIcon, CircleDotIcon, CircleIcon, SquareCheckIcon, SquareIcon, XIcon } from "lucide-react"
import { Markdown } from "@/components/Markdown"
import { cn } from "@/lib/utils"
import type { Question } from "../types/quiz"

type Props = {
  question: Question
  selected: number[]
  onToggle: (index: number) => void
}

export function ChoiceOptions({ question, selected, onToggle }: Props) {
  const result = question.result
  const multi = question.kind === "multi_choice"
  const picked = result?.response.selected ?? selected

  return (
    <div className="flex flex-col gap-2">
      {question.options.map((option, i) => {
        const isPicked = picked.includes(i)
        const isCorrect = result?.correctOptions.includes(i)
        const Indicator = multi ? (isPicked ? SquareCheckIcon : SquareIcon) : isPicked ? CircleDotIcon : CircleIcon
        return (
          <button
            key={i}
            type="button"
            disabled={!!result}
            onClick={() => onToggle(i)}
            className={cn(
              "flex items-start gap-3 rounded-lg border px-3 py-2.5 text-left text-sm transition-colors",
              !result && "hover:border-primary/50 hover:bg-primary/5",
              !result && isPicked && "border-primary bg-primary/10",
              result && isCorrect && "border-success/60 bg-success/10",
              result && isPicked && !isCorrect && "border-destructive/60 bg-destructive/10",
            )}
          >
            <Indicator className={cn("mt-0.5 size-4 shrink-0 text-muted-foreground", !result && isPicked && "text-primary")} />
            <Markdown className="prose-sm min-w-0 flex-1 prose-p:my-0">{option}</Markdown>
            {result && isCorrect && <CheckIcon className="mt-0.5 size-4 shrink-0 text-success" />}
            {result && isPicked && !isCorrect && <XIcon className="mt-0.5 size-4 shrink-0 text-destructive" />}
          </button>
        )
      })}
    </div>
  )
}
