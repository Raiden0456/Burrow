import { useEffect, useRef } from "react"
import { quizApi } from "../api"
import type { QuizResponse } from "../types/quiz"

export function useDraftSaver(quizId: number, questionId: number, response: QuizResponse, enabled: boolean) {
  const last = useRef(JSON.stringify(response))

  useEffect(() => {
    if (!enabled) return
    const serialized = JSON.stringify(response)
    if (serialized === last.current) return
    const timer = setTimeout(() => {
      last.current = serialized
      quizApi.saveDraft(quizId, questionId, response).catch(() => {})
    }, 600)
    return () => clearTimeout(timer)
  }, [quizId, questionId, response, enabled])
}
