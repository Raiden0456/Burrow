import { useCallback, useEffect, useRef, useState } from "react"
import { useQuery, useQueryClient } from "@tanstack/react-query"
import { errorMessage } from "@/lib/api"
import { chatApi } from "../api"
import type { ChatMessage, ThreadDetail } from "../types/chat"

export function useChatThread(id: number) {
  const queryClient = useQueryClient()
  const key = ["chat", "thread", id]
  const thread = useQuery({ queryKey: key, queryFn: () => chatApi.thread(id) })
  const [pending, setPending] = useState<string | null>(null)
  const [reply, setReply] = useState("")
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)
  const abort = useRef<AbortController | null>(null)

  useEffect(() => () => abort.current?.abort(), [])

  const append = useCallback(
    (...messages: ChatMessage[]) =>
      queryClient.setQueryData<ThreadDetail>(["chat", "thread", id], (prev) => prev && { ...prev, messages: [...prev.messages, ...messages] }),
    [queryClient, id],
  )

  const send = useCallback(
    async (content: string) => {
      const controller = new AbortController()
      abort.current = controller
      setPending(content)
      setReply("")
      setError(null)
      setBusy(true)
      try {
        await chatApi.send(
          id,
          content,
          {
            onUser: (m) => {
              append(m)
              setPending(null)
            },
            onDelta: (text) => setReply((r) => r + text),
            onDone: (m, title) => {
              append(m)
              setReply("")
              queryClient.setQueryData<ThreadDetail>(["chat", "thread", id], (prev) => prev && { ...prev, thread: { ...prev.thread, title } })
              queryClient.invalidateQueries({ queryKey: ["chat", "threads"] })
            },
          },
          controller.signal,
        )
      } catch (err) {
        if (!controller.signal.aborted) setError(errorMessage(err))
      } finally {
        setPending(null)
        setBusy(false)
      }
    },
    [id, append, queryClient],
  )

  return { thread, send, pending, reply, error, streaming: busy }
}
