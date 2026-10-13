import { useCallback, useEffect, useRef, useState } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { errorMessage } from "@/lib/api"
import { quizApi } from "../api"
import type { HintKind, Quiz } from "../types/quiz"

export function useHintStream(quizId: number, questionId: number) {
  const queryClient = useQueryClient()
  const [text, setText] = useState("")
  const [kind, setKind] = useState<HintKind | null>(null)
  const [error, setError] = useState<string | null>(null)
  const abort = useRef<AbortController | null>(null)

  useEffect(() => () => abort.current?.abort(), [])

  const start = useCallback(
    async (next: HintKind) => {
      const controller = new AbortController()
      abort.current = controller
      setKind(next)
      setText("")
      setError(null)
      try {
        await quizApi.hint(
          quizId,
          questionId,
          next,
          (delta) => setText((t) => t + delta),
          (hint) =>
            queryClient.setQueryData<Quiz>(["quiz", quizId], (prev) =>
              prev && { ...prev, questions: prev.questions.map((q) => (q.id === questionId ? { ...q, hints: [...q.hints, hint] } : q)) },
            ),
          controller.signal,
        )
      } catch (err) {
        if (!controller.signal.aborted) setError(errorMessage(err))
      } finally {
        if (abort.current === controller) {
          setKind(null)
          setText("")
        }
      }
    },
    [quizId, questionId, queryClient],
  )

  return { text, streaming: kind, error, start }
}
