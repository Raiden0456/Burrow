import { useQuery } from "@tanstack/react-query"
import { Link } from "react-router"
import { CodeIcon, FlameIcon, SparklesIcon, TargetIcon, ZapIcon, CalendarDaysIcon } from "lucide-react"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { PageLoader } from "@/components/PageLoader"
import { ErrorAlert } from "@/components/ErrorAlert"
import { IconChip } from "@/components/IconChip"
import { ActivityGrid, StatTile, useProgress } from "@/features/progress"
import { settingsApi } from "@/features/settings"
import { useCurriculum } from "@/features/topics"
import { useI18n } from "@/i18n"
import { ContinueCard } from "./ContinueCard"
import { DailyCard } from "./DailyCard"
import { HeroCard } from "./HeroCard"
import { ModulesCard } from "./ModulesCard"
import { RecentQuizzes } from "./RecentQuizzes"
import { ReviewCard } from "./ReviewCard"
import { UnfinishedQuizzes } from "./UnfinishedQuizzes"

export function DashboardPage() {
  const { t } = useI18n()
  const progress = useProgress()
  const curriculum = useCurriculum()
  const providers = useQuery({ queryKey: ["providers"], queryFn: settingsApi.providers })

  if (progress.isPending || curriculum.isPending) return <PageLoader />
  if (progress.error) return <ErrorAlert error={progress.error} />
  if (curriculum.error) return <ErrorAlert error={curriculum.error} />

  const p = progress.data
  return (
    <div className="flex flex-col gap-6">
      <HeroCard streak={p.streak} activeToday={p.activeToday} />

      {providers.data?.length === 0 && (
        <Alert className="border-hue-yellow/40 bg-hue-yellow/10">
          <SparklesIcon className="text-hue-yellow" />
          <AlertTitle>{t("dashboard.noProvider.title")}</AlertTitle>
          <AlertDescription>
            <span>
              {t("dashboard.noProvider.body")}{" "}
              <Link to="/settings" className="font-medium text-primary underline underline-offset-2">
                {t("dashboard.noProvider.link")}
              </Link>
            </span>
          </AlertDescription>
        </Alert>
      )}

      <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
        <StatTile icon={FlameIcon} tone="orange" label={t("dashboard.stat.streak")} value={p.streak} hint={t("dashboard.stat.streakDays", { count: p.streak })} />
        <StatTile icon={ZapIcon} tone="yellow" label={t("dashboard.stat.answersToday")} value={p.answersToday} hint={t("dashboard.stat.total", { count: p.totalAnswers })} />
        <StatTile icon={TargetIcon} tone="green" label={t("dashboard.stat.accuracy")} value={p.accuracy7d === null ? "—" : `${Math.round(p.accuracy7d * 100)}%`} />
        <StatTile icon={CodeIcon} tone="purple" label={t("dashboard.stat.drills")} value={p.drillsSolved} />
      </div>

      <UnfinishedQuizzes />

      <div className="grid gap-6 md:grid-cols-2">
        <DailyCard />
        <ReviewCard due={p.dueCount} />
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <ContinueCard modules={curriculum.modules} stats={p.topics} />
        <Card>
          <CardHeader className="flex flex-row items-center gap-3">
            <IconChip icon={CalendarDaysIcon} tone="green" />
            <CardTitle>{t("dashboard.activity.title")}</CardTitle>
          </CardHeader>
          <CardContent>
            <ActivityGrid days={p.activity} />
          </CardContent>
        </Card>
      </div>

      <ModulesCard modules={curriculum.modules} stats={p.topics} />
      <RecentQuizzes />
    </div>
  )
}
