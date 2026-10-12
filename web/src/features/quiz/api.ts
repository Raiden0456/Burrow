import { api } from "@/lib/api"
import type { AnswerOutcome, Quiz, QuizResponse, QuizSummary } from "./types/quiz"

type Created = { id: number }

export const quizApi = {
  createForTopic: (slug: string, size: number) => api.post<Created>(`/api/topics/${slug}/quizzes`, { size }),
  createReview: (size: number) => api.post<Created>("/api/quizzes/review", { size }),
  list: (limit = 10) => api.get<QuizSummary[]>(`/api/quizzes?limit=${limit}`),
  get: (id: number) => api.get<Quiz>(`/api/quizzes/${id}`),
  answer: (quizId: number, questionId: number, response: QuizResponse) =>
    api.post<AnswerOutcome>(`/api/quizzes/${quizId}/answers`, { questionId, response }),
}
