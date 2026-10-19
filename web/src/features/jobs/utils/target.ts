import type { Job } from "../types/job"

export function jobTarget(job: Job): string | null {
  if (job.status !== "done") return null
  switch (job.kind) {
    case "lesson":
      return `/topics/${job.topicSlug}`
    case "drill":
      return job.resultId ? `/drills/${job.resultId}` : null
    case "quiz":
    case "checkpoint":
    case "daily":
      return job.resultId ? `/quiz/${job.resultId}` : null
  }
}
