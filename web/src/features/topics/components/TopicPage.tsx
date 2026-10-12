import { Link, useParams } from "react-router"
import { BookOpenTextIcon, ChevronLeftIcon, DumbbellIcon, TargetIcon } from "lucide-react"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { PageHeader } from "@/components/PageHeader"
import { PageLoader } from "@/components/PageLoader"
import { ErrorAlert } from "@/components/ErrorAlert"
import { MasteryBar, useProgress } from "@/features/progress"
import { useI18n } from "@/i18n"
import { useCurriculum } from "../hooks/useCurriculum"
import { LessonPanel } from "./LessonPanel"
import { PracticePanel } from "./PracticePanel"

export function TopicPage() {
  const { t } = useI18n()
  const { slug = "" } = useParams()
  const curriculum = useCurriculum()
  const { byTopic } = useProgress()

  if (curriculum.isPending) return <PageLoader />
  if (curriculum.error) return <ErrorAlert error={curriculum.error} />
  const topic = curriculum.topics.get(slug)
  if (!topic) return <ErrorAlert error={t("errors.unknownTopic", { slug })} />

  return (
    <>
      <Link to="/topics" className="mb-3 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground">
        <ChevronLeftIcon className="size-4" />
        {topic.module}
      </Link>
      <PageHeader title={topic.title} />
      <div className="mb-6 flex flex-col gap-3 rounded-xl border bg-gradient-to-br from-hue-blue/10 to-hue-purple/5 p-4">
        <MasteryBar value={byTopic.get(slug)?.mastery ?? 0} />
        <ul className="space-y-1.5 text-sm text-muted-foreground">
          {topic.objectives.map((o) => (
            <li key={o} className="flex gap-2">
              <TargetIcon className="mt-0.5 size-3.5 shrink-0 text-hue-blue" />
              {o}
            </li>
          ))}
        </ul>
      </div>
      <Tabs defaultValue="lesson">
        <TabsList>
          <TabsTrigger value="lesson">
            <BookOpenTextIcon />
            {t("topics.tab.lesson")}
          </TabsTrigger>
          <TabsTrigger value="practice">
            <DumbbellIcon />
            {t("topics.tab.practice")}
          </TabsTrigger>
        </TabsList>
        <TabsContent value="lesson" className="pt-4">
          <LessonPanel slug={slug} />
        </TabsContent>
        <TabsContent value="practice" className="pt-4">
          <PracticePanel topic={topic} />
        </TabsContent>
      </Tabs>
    </>
  )
}
