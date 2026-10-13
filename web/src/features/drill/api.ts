import type { Job } from "@/features/jobs"
import { api, stream } from "@/lib/api"
import type { Drill, DrillHint, DrillSolution, DrillSummary, RunResult } from "./types/drill"

export const drillApi = {
  listForTopic: (slug: string) => api.get<DrillSummary[]>(`/api/topics/${slug}/drills`),
  create: (slug: string) => api.post<Job>(`/api/topics/${slug}/drills`),
  get: (id: number) => api.get<Drill>(`/api/drills/${id}`),
  solution: (id: number) => api.get<DrillSolution>(`/api/drills/${id}/solution`),
  run: (id: number, code: string) => api.post<RunResult>(`/api/drills/${id}/run`, { code }),
  saveDraft: (id: number, code: string) => api.put<void>(`/api/drills/${id}/draft`, { code }),
  help: (id: number, code: string, output: string, onDelta: (text: string) => void, onDone: (hint: DrillHint) => void, signal?: AbortSignal) =>
    stream(`/api/drills/${id}/help`, { code, output }, { delta: (d: { text: string }) => onDelta(d.text), done: (h: DrillHint) => onDone(h) }, signal),
}
