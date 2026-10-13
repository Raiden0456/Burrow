import { Loader2Icon } from "lucide-react"
import { useI18n } from "@/i18n"
import { cn } from "@/lib/utils"
import { useJobs } from "../hooks/useJobs"
import { useJobLabel } from "../hooks/useJobLabel"

export function JobsIndicator({ className }: { className?: string }) {
  const { t } = useI18n()
  const jobs = useJobs()
  const label = useJobLabel()
  const running = jobs.data?.filter((j) => j.status === "running") ?? []
  if (running.length === 0) return null

  return (
    <div className={cn("flex flex-col gap-1.5 rounded-lg border border-hue-blue/30 bg-hue-blue/8 p-2.5 text-xs", className)}>
      <div className="font-medium text-hue-blue">{t("jobs.running", { count: running.length })}</div>
      {running.map((job) => (
        <div key={job.id} className="flex items-center gap-1.5 text-muted-foreground">
          <Loader2Icon className="size-3 shrink-0 animate-spin text-hue-blue" />
          <span className="truncate">{label(job)}</span>
        </div>
      ))}
    </div>
  )
}
