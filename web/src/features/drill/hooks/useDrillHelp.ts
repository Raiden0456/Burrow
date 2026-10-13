import { useCallback, useEffect, useRef, useState } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { errorMessage } from "@/lib/api"
import { drillApi } from "../api"
import type { Drill } from "../types/drill"

export function useDrillHelp(id: number) {
  const queryClient = useQueryClient()
  const [text, setText] = useState("")
  const [streaming, setStreaming] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const abort = useRef<AbortController | null>(null)

  useEffect(() => () => abort.current?.abort(), [])

  const ask = useCallback(
    async (code: string, output: string) => {
      const controller = new AbortController()
      abort.current = controller
      setText("")
      setError(null)
      setStreaming(true)
      try {
        await drillApi.help(
          id,
          code,
          output,
          (delta) => setText((t) => t + delta),
          (hint) => queryClient.setQueryData<Drill>(["drill", id], (prev) => prev && { ...prev, hints: [...prev.hints, hint] }),
          controller.signal,
        )
      } catch (err) {
        if (!controller.signal.aborted) setError(errorMessage(err))
      } finally {
        if (abort.current === controller) {
          setStreaming(false)
          setText("")
        }
      }
    },
    [id, queryClient],
  )

  return { text, streaming, error, ask }
}
