import { api } from "@/lib/api"
import type { Drill, DrillSolution, DrillSummary, RunResult } from "./types/drill"

export const drillApi = {
  listForTopic: (slug: string) => api.get<DrillSummary[]>(`/api/topics/${slug}/drills`),
  create: (slug: string) => api.post<{ id: number }>(`/api/topics/${slug}/drills`),
  get: (id: number) => api.get<Drill>(`/api/drills/${id}`),
  solution: (id: number) => api.get<DrillSolution>(`/api/drills/${id}/solution`),
  run: (id: number, code: string) => api.post<RunResult>(`/api/drills/${id}/run`, { code }),
}
