import { Link } from "react-router"
import { BadgeCheckIcon, CheckCircle2Icon, CodeIcon } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { PageHeader } from "@/components/PageHeader"
import { PageLoader } from "@/components/PageLoader"
import { ErrorAlert } from "@/components/ErrorAlert"
import { MasteryBar, useProgress } from "@/features/progress"
import { useI18n } from "@/i18n"
import { useCurriculum } from "../hooks/useCurriculum"

export function TopicsPage() {
  const { t } = useI18n()
  const curriculum = useCurriculum()
  const { byTopic } = useProgress()

  if (curriculum.isPending) return <PageLoader />
  if (curriculum.error) return <ErrorAlert error={curriculum.error} />

  return (
    <>
      <PageHeader title={t("topics.title")} description={t("topics.description")} />
      <div className="flex flex-col gap-6">
        {curriculum.modules.map((module, mi) => (
          <Card key={module.slug}>
            <CardHeader className="flex flex-row items-center gap-3">
              <span className="flex size-8 items-center justify-center rounded-lg bg-primary/12 text-sm font-semibold text-primary">{mi + 1}</span>
              <CardTitle>{module.title}</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col divide-y">
              {module.topics.map((topic) => {
                const stat = byTopic.get(topic.slug)
                const mastered = (stat?.mastery ?? 0) >= 0.7
                return (
                  <Link
                    key={topic.slug}
                    to={`/topics/${topic.slug}`}
                    className="-mx-2 flex flex-col gap-2 rounded-md px-2 py-3 transition-colors hover:bg-muted/60 sm:flex-row sm:items-center sm:justify-between"
                  >
                    <div className="min-w-0">
                      <div className="flex items-center gap-2 font-medium">
                        {topic.title}
                        {mastered && (
                          <span className="inline-flex items-center gap-1 rounded-full bg-success/15 px-2 py-0.5 text-[11px] font-medium text-success">
                            <BadgeCheckIcon className="size-3" />
                            {t("topics.mastered")}
                          </span>
                        )}
                      </div>
                      <div className="mt-0.5 flex gap-3 text-xs text-muted-foreground">
                        <span className="flex items-center gap-1">
                          <CheckCircle2Icon className="size-3 text-hue-green" />
                          {t("topics.answered", { count: stat?.answered ?? 0 })}
                        </span>
                        <span className="flex items-center gap-1">
                          <CodeIcon className="size-3 text-hue-purple" />
                          {t("topics.drills", { count: stat?.drillsSolved ?? 0 })}
                        </span>
                      </div>
                    </div>
                    <MasteryBar value={stat?.mastery ?? 0} className="w-full sm:w-40" />
                  </Link>
                )
              })}
            </CardContent>
          </Card>
        ))}
      </div>
    </>
  )
}
