import { useQuery } from "@tanstack/react-query"
import { Link } from "react-router"
import { BadgeCheckIcon, CheckCircle2Icon, CodeIcon } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { PageHeader } from "@/components/PageHeader"
import { PageLoader } from "@/components/PageLoader"
import { ErrorAlert } from "@/components/ErrorAlert"
import { useJobs } from "@/features/jobs"
import { MasteryBar, useProgress } from "@/features/progress"
import { useI18n } from "@/i18n"
import { enter } from "@/lib/enter"
import { topicsApi } from "../api"
import { useCurriculum } from "../hooks/useCurriculum"
import { LessonAction } from "./LessonAction"

export function TopicsPage() {
  const { t } = useI18n()
  const curriculum = useCurriculum()
  const { byTopic } = useProgress()
  const lessons = useQuery({ queryKey: ["lessons"], queryFn: topicsApi.lessons })
  const jobs = useJobs()

  if (curriculum.isPending) return <PageLoader />
  if (curriculum.error) return <ErrorAlert error={curriculum.error} />

  const withLesson = new Set(lessons.data?.map((l) => l.topicSlug))
  const generating = new Set(jobs.data?.filter((j) => j.kind === "lesson" && j.status === "running").map((j) => j.topicSlug))

  return (
    <>
      <PageHeader title={t("topics.title")} description={t("topics.description")} />
      <div className="flex flex-col gap-6">
        {curriculum.modules.map((module, mi) => (
          <Card key={module.slug} {...enter(mi + 1)}>
            <CardHeader className="flex flex-row items-center gap-3">
              <span className="flex size-8 items-center justify-center rounded-lg bg-primary/12 text-sm font-semibold text-primary">{mi + 1}</span>
              <CardTitle>{module.title}</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col divide-y">
              {module.topics.map((topic) => {
                const stat = byTopic.get(topic.slug)
                const mastered = (stat?.mastery ?? 0) >= 0.7
                return (
                  <div key={topic.slug} className="-mx-2 flex flex-col gap-2 rounded-md px-2 py-3 transition-colors hover:bg-muted/60 sm:flex-row sm:items-center">
                    <Link to={`/topics/${topic.slug}`} className="min-w-0 flex-1">
                      <div className="flex flex-wrap items-center gap-2 font-medium">
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
                    </Link>
                    <div className="flex items-center gap-3">
                      <LessonAction slug={topic.slug} hasLesson={withLesson.has(topic.slug)} generating={generating.has(topic.slug)} />
                      <Link to={`/topics/${topic.slug}`} className="w-full sm:w-40">
                        <MasteryBar value={stat?.mastery ?? 0} />
                      </Link>
                    </div>
                  </div>
                )
              })}
            </CardContent>
          </Card>
        ))}
      </div>
    </>
  )
}
