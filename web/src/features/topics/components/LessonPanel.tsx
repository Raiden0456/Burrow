import { useQuery, useQueryClient } from "@tanstack/react-query"
import { useEffect, useRef } from "react"
import { Loader2Icon, RefreshCwIcon, SparklesIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { Markdown } from "@/components/Markdown"
import { ErrorAlert } from "@/components/ErrorAlert"
import { PageLoader } from "@/components/PageLoader"
import { useI18n } from "@/i18n"
import { topicsApi } from "../api"
import { useGenerateLesson } from "../hooks/useGenerateLesson"

export function LessonPanel({ slug }: { slug: string }) {
  const { t } = useI18n()
  const queryClient = useQueryClient()
  const lesson = useQuery({
    queryKey: ["lesson", slug],
    queryFn: () => topicsApi.lesson(slug),
    refetchInterval: (q) => (q.state.data?.job ? 1200 : false),
  })
  const generate = useGenerateLesson()

  const running = lesson.data?.job ?? null
  const wasRunning = useRef(false)
  useEffect(() => {
    if (wasRunning.current && !running) queryClient.invalidateQueries({ queryKey: ["jobs"] })
    wasRunning.current = !!running
  }, [running, queryClient])

  if (lesson.isPending) return <PageLoader />
  if (lesson.error) return <ErrorAlert error={lesson.error} />

  const start = () => generate.mutate(slug)
  const content = running ? running.partial || "" : lesson.data.content

  return (
    <div className="flex flex-col gap-4">
      {running && (
        <div className="flex items-center gap-2 rounded-lg border border-hue-blue/30 bg-hue-blue/8 px-3 py-2 text-sm text-hue-blue">
          <Loader2Icon className="size-4 animate-spin" />
          {t("topics.lesson.background")}
        </div>
      )}

      {content ? (
        <>
          <div data-glass className="wallpaper:rounded-xl wallpaper:p-5">
            <Markdown>{content}</Markdown>
          </div>
          {running && <Skeleton className="h-4 w-1/3" />}
          {!running && (
            <div>
              <Button variant="outline" size="sm" onClick={start} disabled={generate.isPending}>
                <RefreshCwIcon />
                {t("topics.lesson.regenerate")}
              </Button>
            </div>
          )}
        </>
      ) : running ? (
        <div className="flex flex-col gap-2">
          <Skeleton className="h-6 w-1/2" />
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-5/6" />
        </div>
      ) : (
        <div className="flex flex-col items-start gap-3 rounded-xl border border-dashed border-hue-blue/40 bg-hue-blue/5 p-6">
          <p className="text-sm text-muted-foreground">{t("topics.lesson.empty")}</p>
          <Button onClick={start} disabled={generate.isPending} className="max-md:w-full">
            <SparklesIcon />
            {t("topics.lesson.generate")}
          </Button>
        </div>
      )}
    </div>
  )
}
