import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Link, useNavigate } from "react-router"
import { CalendarCheckIcon, Loader2Icon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { ErrorAlert } from "@/components/ErrorAlert"
import { IconChip } from "@/components/IconChip"
import { useJob, useRunningJob } from "@/features/jobs"
import { useI18n } from "@/i18n"
import { dashboardApi } from "../api"

export function DailyCard() {
  const { t, locale } = useI18n()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [jobId, setJobId] = useState<number | null>(null)
  const [failure, setFailure] = useState<string | null>(null)
  const background = useRunningJob("daily", "")
  const daily = useQuery({ queryKey: ["daily"], queryFn: dashboardApi.daily })

  const job = useJob(jobId ?? background?.id ?? null, (finished) => {
    queryClient.invalidateQueries({ queryKey: ["daily"] })
    queryClient.invalidateQueries({ queryKey: ["jobs"] })
    if (finished.status === "done" && finished.resultId && jobId === finished.id) navigate(`/quiz/${finished.resultId}`)
    if (finished.status === "failed" && jobId === finished.id) setFailure(finished.error)
    setJobId(null)
  })
  const create = useMutation({
    mutationFn: dashboardApi.createDaily,
    onSuccess: (started) => {
      setFailure(null)
      setJobId(started.id)
    },
  })
  const building = create.isPending || job?.status === "running"

  const d = daily.data
  const time = d && new Date(d.scheduledAt * 1000).toLocaleTimeString(locale, { hour: "2-digit", minute: "2-digit" })
  const description = !d ? " " : d.quizId ? t("dashboard.daily.ready") : d.enabled ? t("dashboard.daily.scheduled", { time: time ?? "" }) : t("dashboard.daily.off")

  return (
    <Card className="border-l-4 border-l-hue-blue">
      <CardHeader className="flex flex-row items-start gap-3">
        <IconChip icon={CalendarCheckIcon} tone="blue" />
        <div className="flex flex-col gap-1">
          <CardTitle>{t("dashboard.daily.title")}</CardTitle>
          <CardDescription>{description}</CardDescription>
        </div>
      </CardHeader>
      <CardContent className="flex flex-col gap-2">
        {daily.isPending ? (
          <Skeleton className="h-8 w-32" />
        ) : d?.quizId ? (
          <div>
            <Button asChild className="max-md:w-full">
              <Link to={`/quiz/${d.quizId}`}>{t("dashboard.daily.open")}</Link>
            </Button>
          </div>
        ) : (
          <div>
            <Button onClick={() => create.mutate()} disabled={building} className="max-md:w-full">
              {building && <Loader2Icon className="animate-spin" />}
              {building ? t("dashboard.daily.building") : t("dashboard.daily.start")}
            </Button>
          </div>
        )}
        {building && <p className="text-xs text-muted-foreground">{t("jobs.leaveHint")}</p>}
        {failure && <ErrorAlert error={failure} />}
      </CardContent>
    </Card>
  )
}
