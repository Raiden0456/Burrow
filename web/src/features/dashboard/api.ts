import { api } from "@/lib/api"
import type { DailyStatus } from "./types/daily"

export const dashboardApi = {
  daily: () => api.get<DailyStatus>("/api/daily"),
  createDaily: () => api.post<{ id: number }>("/api/daily"),
}
