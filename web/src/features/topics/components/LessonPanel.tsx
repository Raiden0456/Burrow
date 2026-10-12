import { useQuery } from "@tanstack/react-query"
import { RefreshCwIcon, SparklesIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { Markdown } from "@/components/Markdown"
import { ErrorAlert } from "@/components/ErrorAlert"
import { PageLoader } from "@/components/PageLoader"
import { useI18n } from "@/i18n"
import { topicsApi } from "../api"
import { useLessonStream } from "../hooks/useLessonStream"

export function LessonPanel({ slug }: { slug: string }) {
  const { t } = useI18n()
  const lesson = useQuery({ queryKey: ["lesson", slug], queryFn: () => topicsApi.lesson(slug) })
  const gen = useLessonStream(slug)

  if (lesson.isPending) return <PageLoader />
  if (lesson.error) return <ErrorAlert error={lesson.error} />

  const content = gen.streaming || gen.text ? gen.text : lesson.data.content

  return (
    <div className="flex flex-col gap-4">
      {gen.error && <ErrorAlert error={gen.error} title={t("topics.lesson.failed")} />}

      {content ? (
        <>
          <Markdown>{content}</Markdown>
          {gen.streaming && <Skeleton className="h-4 w-1/3" />}
          {!gen.streaming && (
            <div>
              <Button variant="outline" size="sm" onClick={gen.start}>
                <RefreshCwIcon />
                {t("topics.lesson.regenerate")}
              </Button>
            </div>
          )}
        </>
      ) : gen.streaming ? (
        <div className="flex flex-col gap-2">
          <Skeleton className="h-6 w-1/2" />
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-5/6" />
        </div>
      ) : (
        <div className="flex flex-col items-start gap-3 rounded-xl border border-dashed border-hue-blue/40 bg-hue-blue/5 p-6">
          <p className="text-sm text-muted-foreground">{t("topics.lesson.empty")}</p>
          <Button onClick={gen.start}>
            <SparklesIcon />
            {t("topics.lesson.generate")}
          </Button>
        </div>
      )}
    </div>
  )
}
