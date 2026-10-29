import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Link, useNavigate } from "react-router"
import { BugIcon, CheckCircle2Icon, CircleIcon, CodeXmlIcon, Loader2Icon, PlusIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { DifficultyBadge } from "@/components/DifficultyBadge"
import { ErrorAlert } from "@/components/ErrorAlert"
import { IconChip } from "@/components/IconChip"
import { useJob, useRunningJob } from "@/features/jobs"
import { useI18n } from "@/i18n"
import { drillApi } from "../api"
import type { DrillMode } from "../types/drill"

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
    mutationFn: (mode: DrillMode) => drillApi.create(slug, mode),
    onSuccess: (started) => {
      setFailure(null)
      setJobId(started.id)
      queryClient.invalidateQueries({ queryKey: ["jobs"] })
    },
  })
  const generating = create.isPending || job?.status === "running"

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h3 className="flex items-center gap-3 font-medium">
          <IconChip icon={CodeXmlIcon} tone="purple" />
          {t("drills.title")}
        </h3>
        <div className="flex flex-wrap gap-2">
          <Button size="sm" variant="outline" onClick={() => create.mutate("implement")} disabled={generating}>
            <PlusIcon />
            {t("drills.new")}
          </Button>
          <Button size="sm" variant="outline" onClick={() => create.mutate("bug_hunt")} disabled={generating}>
            <BugIcon />
            {t("drills.bugHunt.new")}
          </Button>
        </div>
      </div>
      <p className="text-sm text-muted-foreground">{t("drills.bugHunt.description")}</p>
      {generating && (
        <p className="flex items-center gap-2 text-xs text-muted-foreground">
          <Loader2Icon className="size-3 animate-spin" />
          {t("drills.generating")} {t("jobs.leaveHint")}
        </p>
      )}
      {create.error && <ErrorAlert error={create.error} />}
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
              {d.mode === "bug_hunt" && (
                <span className="flex shrink-0 items-center gap-1 rounded-full bg-hue-yellow/10 px-2 py-0.5 text-xs text-hue-yellow">
                  <BugIcon className="size-3" />
                  {t("drills.bugHunt.title")}
                </span>
              )}
              <DifficultyBadge level={d.difficulty} />
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
