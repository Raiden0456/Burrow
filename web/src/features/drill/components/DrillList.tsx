import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Link, useNavigate } from "react-router"
import { CheckCircle2Icon, CircleIcon, CodeXmlIcon, Loader2Icon, PlusIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { DifficultyBadge } from "@/components/DifficultyBadge"
import { ErrorAlert } from "@/components/ErrorAlert"
import { IconChip } from "@/components/IconChip"
import { useJob, useRunningJob } from "@/features/jobs"
import { useI18n } from "@/i18n"
import { drillApi } from "../api"

export function DrillList({ slug }: { slug: string }) {
  const { t } = useI18n()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [jobId, setJobId] = useState<number | null>(null)
  const [failure, setFailure] = useState<string | null>(null)
  const background = useRunningJob("drill", slug)
  const drills = useQuery({ queryKey: ["drills", slug], queryFn: () => drillApi.listForTopic(slug) })

  const job = useJob(jobId ?? background?.id ?? null, (finished) => {
    queryClient.invalidateQueries({ queryKey: ["drills", slug] })
    queryClient.invalidateQueries({ queryKey: ["jobs"] })
    if (finished.status === "done" && finished.resultId && jobId === finished.id) navigate(`/drills/${finished.resultId}`)
    if (finished.status === "failed" && jobId === finished.id) setFailure(finished.error)
    setJobId(null)
  })

  const create = useMutation({
    mutationFn: () => drillApi.create(slug),
    onSuccess: (started) => {
      setFailure(null)
      setJobId(started.id)
      queryClient.invalidateQueries({ queryKey: ["jobs"] })
    },
  })
  const generating = create.isPending || job?.status === "running"

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between gap-2">
        <h3 className="flex items-center gap-3 font-medium">
          <IconChip icon={CodeXmlIcon} tone="purple" />
          {t("drills.title")}
        </h3>
        <Button size="sm" variant="outline" onClick={() => create.mutate()} disabled={generating}>
          {generating ? <Loader2Icon className="animate-spin" /> : <PlusIcon />}
          {generating ? t("drills.generating") : t("drills.new")}
        </Button>
      </div>
      {generating && <p className="text-xs text-muted-foreground">{t("jobs.leaveHint")}</p>}
      {failure && <ErrorAlert error={failure} />}
      {drills.isPending && <Skeleton className="h-10 w-full" />}
      {drills.error && <ErrorAlert error={drills.error} />}
      {drills.data?.length === 0 && !generating && <p className="text-sm text-muted-foreground">{t("drills.empty")}</p>}
      {!!drills.data?.length && (
        <div data-glass className="flex flex-col divide-y overflow-hidden rounded-lg border">
          {drills.data.map((d) => (
            <Link key={d.id} to={`/drills/${d.id}`} className="flex items-center gap-3 px-3 py-2.5 text-sm transition-colors hover:bg-muted/60 pointer-coarse:py-3.5">
              {d.solved ? <CheckCircle2Icon className="size-4 text-success" /> : <CircleIcon className="size-4 text-muted-foreground" />}
              <span className="min-w-0 flex-1 truncate">{d.title}</span>
              <DifficultyBadge level={d.difficulty} />
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
