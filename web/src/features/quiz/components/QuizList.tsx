import { useQuery } from "@tanstack/react-query"
import { Link } from "react-router"
import { ArrowRightIcon } from "lucide-react"
import { Progress } from "@/components/ui/progress"
import { Skeleton } from "@/components/ui/skeleton"
import { useCurriculum } from "@/features/topics"
import { useI18n } from "@/i18n"
import { quizApi, type QuizFilter } from "../api"
import { QuizKindBadge } from "./QuizKindBadge"

export function QuizList({ filter, empty }: { filter: QuizFilter; empty?: string }) {
  const { t, locale } = useI18n()
  const { topics } = useCurriculum()
  const quizzes = useQuery({ queryKey: ["quizzes", filter], queryFn: () => quizApi.list(filter) })

  if (quizzes.isPending) return <Skeleton className="h-16 w-full" />
  if (!quizzes.data?.length) return empty ? <p className="text-sm text-muted-foreground">{empty}</p> : null

  return (
    <div data-glass className="flex flex-col divide-y overflow-hidden rounded-lg border">
      {quizzes.data.map((q) => {
        const open = q.completedAt === null
        const title = q.topicSlug
          ? (topics.get(q.topicSlug)?.title ?? q.topicSlug)
          : new Date(q.createdAt * 1000).toLocaleDateString(locale, { day: "numeric", month: "long" })
        return (
          <Link key={q.id} to={`/quiz/${q.id}`} className="flex items-center gap-3 px-3 py-2.5 text-sm transition-colors hover:bg-muted/60 pointer-coarse:py-3.5">
            <QuizKindBadge kind={q.kind} />
            <div className="min-w-0 flex-1">
              <div className="truncate">{title}</div>
              {open && <Progress value={(q.answered / Math.max(q.total, 1)) * 100} className="mt-1.5 h-1" />}
            </div>
            <span className="shrink-0 text-xs text-muted-foreground tabular-nums">
              {q.answered}/{q.total}
              {!open && q.score !== null && ` · ${Math.round(q.score * 100)}%`}
            </span>
            {open && (
              <span className="flex shrink-0 items-center gap-1 text-xs font-medium text-primary">
                {t("quiz.continue")}
                <ArrowRightIcon className="size-3.5" />
              </span>
            )}
          </Link>
        )
      })}
    </div>
  )
}
