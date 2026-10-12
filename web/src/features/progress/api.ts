import { api } from "@/lib/api"
import type { ProgressSummary } from "./types/progress"

export const progressApi = {
  summary: () => api.get<ProgressSummary>("/api/progress"),
}
