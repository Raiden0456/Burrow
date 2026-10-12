import { api, stream } from "@/lib/api"
import type { CurriculumModule, Lesson } from "./types/curriculum"

export const topicsApi = {
  curriculum: () => api.get<CurriculumModule[]>("/api/curriculum"),
  lesson: (slug: string) => api.get<Lesson>(`/api/topics/${slug}/lesson`),
  generateLesson: (slug: string, onDelta: (text: string) => void, signal?: AbortSignal) =>
    stream(`/api/topics/${slug}/lesson`, {}, { delta: (d: { text: string }) => onDelta(d.text) }, signal),
}
