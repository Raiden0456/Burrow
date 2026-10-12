import { api, stream } from "@/lib/api"
import type { ChatMessage, Thread, ThreadDetail } from "./types/chat"

type SendHandlers = {
  onUser: (m: ChatMessage) => void
  onDelta: (text: string) => void
  onDone: (m: ChatMessage, title: string) => void
}

export const chatApi = {
  threads: () => api.get<Thread[]>("/api/chat/threads"),
  thread: (id: number) => api.get<ThreadDetail>(`/api/chat/threads/${id}`),
  create: (input: { title?: string; context?: string }) => api.post<Thread>("/api/chat/threads", input),
  remove: (id: number) => api.delete(`/api/chat/threads/${id}`),
  send: (id: number, content: string, h: SendHandlers, signal?: AbortSignal) =>
    stream(
      `/api/chat/threads/${id}/messages`,
      { content },
      {
        user: (m: ChatMessage) => h.onUser(m),
        delta: (d: { text: string }) => h.onDelta(d.text),
        done: (d: { message: ChatMessage; title: string }) => h.onDone(d.message, d.title),
      },
      signal,
    ),
}
