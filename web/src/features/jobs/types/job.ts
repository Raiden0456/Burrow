export type JobKind = "lesson" | "quiz" | "drill" | "daily"

export type JobStatus = "running" | "done" | "failed"

export type Job = {
  id: number
  kind: JobKind
  topicSlug: string
  status: JobStatus
  resultId: number | null
  error: string
  createdAt: number
  finishedAt: number | null
  partial?: string
}
