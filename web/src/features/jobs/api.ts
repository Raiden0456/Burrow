import { api } from "@/lib/api"
import type { Job } from "./types/job"

export const jobsApi = {
  recent: () => api.get<Job[]>("/api/jobs"),
  get: (id: number) => api.get<Job>(`/api/jobs/${id}`),
}
