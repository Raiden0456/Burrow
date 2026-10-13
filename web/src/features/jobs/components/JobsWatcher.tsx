import { useEffect } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { useNavigate } from "react-router"
import { toast } from "sonner"
import { useCurriculum } from "@/features/topics"
import { useI18n } from "@/i18n"
import { useJobs } from "../hooks/useJobs"
import { jobTarget } from "../utils/target"
import { markNotified, wasNotified } from "../utils/notified"
import { useJobLabel } from "../hooks/useJobLabel"

export function JobsWatcher() {
  const { t } = useI18n()
  const jobs = useJobs()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { isPending } = useCurriculum()
  const label = useJobLabel()

  useEffect(() => {
    if (isPending) return
    for (const job of jobs.data ?? []) {
      if (job.status === "running" || wasNotified(job.id)) continue
      markNotified(job.id)

      queryClient.invalidateQueries({ queryKey: ["lesson", job.topicSlug] })
      queryClient.invalidateQueries({ queryKey: ["lessons"] })
      queryClient.invalidateQueries({ queryKey: ["drills", job.topicSlug] })
      queryClient.invalidateQueries({ queryKey: ["quizzes"] })
      queryClient.invalidateQueries({ queryKey: ["daily"] })

      const target = jobTarget(job)
      if (job.status === "done") {
        toast.success(t("jobs.ready", { what: label(job) }), {
          duration: 10_000,
          action: target ? { label: t("jobs.open"), onClick: () => navigate(target) } : undefined,
        })
      } else {
        toast.error(t("jobs.failed", { what: label(job) }), { description: job.error, duration: 15_000 })
      }
    }
  }, [jobs.data, isPending, queryClient, navigate, t, label])

  return null
}
