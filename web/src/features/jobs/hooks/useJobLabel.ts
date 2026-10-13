import { useCallback } from "react"
import { useCurriculum } from "@/features/topics"
import { useI18n } from "@/i18n"
import type { Job } from "../types/job"

export function useJobLabel() {
  const { t } = useI18n()
  const { topics } = useCurriculum()
  return useCallback(
    (job: Job) => {
      const topic = topics.get(job.topicSlug)?.title ?? job.topicSlug
      return job.kind === "daily" ? t("jobs.kind.daily") : t(`jobs.kind.${job.kind}`, { topic })
    },
    [t, topics],
  )
}
