import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useNavigate } from "react-router"
import { chatApi } from "../api"

export type StartChatInput = {
  title?: string
  context?: string
  message?: string
}

export function useStartChat() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ title, context }: StartChatInput) => chatApi.create({ title, context }),
    onSuccess: (thread, { message }) => {
      queryClient.invalidateQueries({ queryKey: ["chat", "threads"] })
      navigate(`/chat/${thread.id}`, { state: message ? { autoSend: message } : undefined })
    },
  })
}
