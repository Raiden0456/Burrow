import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useNavigate } from "react-router"
import { BadgeCheckIcon, FlagIcon, Loader2Icon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { ErrorAlert } from "@/components/ErrorAlert"
import { IconChip } from "@/components/IconChip"
import { useJob, useRunningJob } from "@/features/jobs"
import { useProgress } from "@/features/progress"
import { quizApi } from "@/features/quiz"
import { useI18n } from "@/i18n"

const checkpointSize = 8

export function CheckpointCard({ slug }: { slug: string }) {
  const { t } = useI18n()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { byTopic } = useProgress()
  const [jobId, setJobId] = useState<number | null>(null)
  const [failure, setFailure] = useState<string | null>(null)
  const background = useRunningJob("checkpoint", slug)

  const job = useJob(jobId ?? background?.id ?? null, (finished) => {
    queryClient.invalidateQueries({ queryKey: ["quizzes"] })
    queryClient.invalidateQueries({ queryKey: ["jobs"] })
    if (finished.status === "done" && finished.resultId && jobId === finished.id) navigate(`/quiz/${finished.resultId}`)
    if (finished.status === "failed" && jobId === finished.id) setFailure(finished.error)
    setJobId(null)
  })

  const start = useMutation({
    mutationFn: () => quizApi.createCheckpoint(slug),
    onSuccess: (started) => {
      setFailure(null)
      setJobId(started.id)
      queryClient.invalidateQueries({ queryKey: ["jobs"] })
    },
  })
  const generating = start.isPending || job?.status === "running"

  if (byTopic.get(slug)?.learned) {
    return (
      <div data-glass className="flex items-center gap-3 rounded-xl border bg-success/10 p-4 text-sm text-success">
        <BadgeCheckIcon className="size-5 shrink-0" />
        {t("topics.checkpoint.mastered")}
      </div>
    )
  }

  return (
    <Card className="border-l-4 border-l-hue-green">
      <CardHeader className="flex flex-row items-start gap-3">
        <IconChip icon={FlagIcon} tone="green" />
        <div className="flex flex-col gap-1">
          <CardTitle>{t("topics.checkpoint.title")}</CardTitle>
          <CardDescription>{t("topics.checkpoint.description", { count: checkpointSize })}</CardDescription>
        </div>
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        <Button variant="outline" className="self-start max-md:self-stretch" onClick={() => start.mutate()} disabled={generating}>
          {generating && <Loader2Icon className="animate-spin" />}
          {generating ? t("topics.quiz.generating") : t("topics.checkpoint.start")}
        </Button>
        {generating && <p className="text-xs text-muted-foreground">{t("jobs.leaveHint")}</p>}
        {failure && <ErrorAlert error={failure} />}
      </CardContent>
    </Card>
  )
}
