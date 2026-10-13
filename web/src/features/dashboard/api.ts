import type { Job } from "@/features/jobs"
import { api } from "@/lib/api"
import type { DailyStatus } from "./types/daily"

export const dashboardApi = {
  daily: () => api.get<DailyStatus>("/api/daily"),
  createDaily: () => api.post<Job>("/api/daily"),
}
