import type { Job } from "@/features/jobs"
import { api, stream } from "@/lib/api"
import type { AnswerOutcome, Hint, HintKind, Quiz, QuizResponse, QuizSummary } from "./types/quiz"

export type QuizFilter = {
  topic?: string
  status?: "open" | "done"
  limit?: number
}

function query(f: QuizFilter) {
  const params = new URLSearchParams()
  if (f.topic) params.set("topic", f.topic)
  if (f.status) params.set("status", f.status)
  params.set("limit", String(f.limit ?? 20))
  return params.toString()
}

export const quizApi = {
  createForTopic: (slug: string, size: number) => api.post<Job>(`/api/topics/${slug}/quizzes`, { size }),
  createCheckpoint: (slug: string) => api.post<Job>(`/api/topics/${slug}/checkpoint`),
  createReview: (size: number) => api.post<{ id: number }>("/api/quizzes/review", { size }),
  list: (f: QuizFilter = {}) => api.get<QuizSummary[]>(`/api/quizzes?${query(f)}`),
  get: (id: number) => api.get<Quiz>(`/api/quizzes/${id}`),
  delete: (id: number) => api.delete(`/api/quizzes/${id}`),
  answer: (quizId: number, questionId: number, response: QuizResponse) =>
    api.post<AnswerOutcome>(`/api/quizzes/${quizId}/answers`, { questionId, response }),
  saveDraft: (quizId: number, questionId: number, response: QuizResponse) =>
    api.put<void>(`/api/quizzes/${quizId}/drafts`, { questionId, response }),
  hint: (quizId: number, questionId: number, kind: HintKind, onDelta: (text: string) => void, onDone: (hint: Hint) => void, signal?: AbortSignal) =>
    stream(
      `/api/quizzes/${quizId}/questions/${questionId}/hints`,
      { kind },
      { delta: (d: { text: string }) => onDelta(d.text), done: (h: Hint) => onDone(h) },
      signal,
    ),
}
