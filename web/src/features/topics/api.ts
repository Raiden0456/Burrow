import type { Job } from "@/features/jobs"
import { api } from "@/lib/api"
import type { CurriculumModule, Lesson, LessonSummary } from "./types/curriculum"

export const topicsApi = {
  curriculum: () => api.get<CurriculumModule[]>("/api/curriculum"),
  lessons: () => api.get<LessonSummary[]>("/api/lessons"),
  lesson: (slug: string) => api.get<Lesson>(`/api/topics/${slug}/lesson`),
  generateLesson: (slug: string) => api.post<Job>(`/api/topics/${slug}/lesson`),
}
