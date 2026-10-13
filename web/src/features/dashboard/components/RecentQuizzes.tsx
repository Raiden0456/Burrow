import { useQuery } from "@tanstack/react-query"
import { Link } from "react-router"
import { HistoryIcon } from "lucide-react"
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { IconChip } from "@/components/IconChip"
import { QuizKindBadge, quizApi } from "@/features/quiz"
import { useCurriculum } from "@/features/topics"
import { useI18n } from "@/i18n"

export function RecentQuizzes() {
  const { t, locale } = useI18n()
  const { topics } = useCurriculum()
  const quizzes = useQuery({ queryKey: ["quizzes", "recent"], queryFn: () => quizApi.list({ limit: 8 }) })
  if (!quizzes.data?.length) return null

  return (
    <Card>
      <CardHeader className="flex flex-row items-center gap-3">
        <IconChip icon={HistoryIcon} tone="cyan" />
        <CardTitle>{t("dashboard.recent.title")}</CardTitle>
        <CardAction>
          <Link to="/quizzes" className="text-sm text-primary hover:underline">
            {t("dashboard.unfinished.all")}
          </Link>
        </CardAction>
      </CardHeader>
      <CardContent className="flex flex-col divide-y">
        {quizzes.data.map((q) => (
          <Link key={q.id} to={`/quiz/${q.id}`} className="flex items-center gap-3 py-2 text-sm hover:text-primary pointer-coarse:py-3">
            <QuizKindBadge kind={q.kind} />
            <span className="min-w-0 flex-1 truncate">
              {q.topicSlug ? (topics.get(q.topicSlug)?.title ?? q.topicSlug) : new Date(q.createdAt * 1000).toLocaleDateString(locale, { day: "numeric", month: "long" })}
            </span>
            <span className="text-xs text-muted-foreground tabular-nums">
              {q.answered}/{q.total}
              {q.score !== null && ` · ${Math.round(q.score * 100)}%`}
            </span>
          </Link>
        ))}
      </CardContent>
    </Card>
  )
}
