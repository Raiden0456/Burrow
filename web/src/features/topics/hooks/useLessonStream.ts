import { useCallback, useEffect, useRef, useState } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { errorMessage } from "@/lib/api"
import { topicsApi } from "../api"

export function useLessonStream(slug: string) {
  const queryClient = useQueryClient()
  const [text, setText] = useState("")
  const [streaming, setStreaming] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const abort = useRef<AbortController | null>(null)

  useEffect(() => () => abort.current?.abort(), [])

  const start = useCallback(async () => {
    abort.current?.abort()
    const controller = new AbortController()
    abort.current = controller
    setText("")
    setError(null)
    setStreaming(true)
    try {
      await topicsApi.generateLesson(slug, (delta) => setText((t) => t + delta), controller.signal)
      await queryClient.invalidateQueries({ queryKey: ["lesson", slug] })
    } catch (err) {
      if (!controller.signal.aborted) setError(errorMessage(err))
    } finally {
      if (abort.current === controller) setStreaming(false)
    }
  }, [slug, queryClient])

  return { text, streaming, error, start }
}
