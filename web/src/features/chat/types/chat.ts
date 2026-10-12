export type Thread = {
  id: number
  title: string
  context: string
  createdAt: number
  updatedAt: number
}

export type ChatRole = "user" | "assistant"

export type ChatMessage = {
  id: number
  role: ChatRole
  content: string
  createdAt: number
}

export type ThreadDetail = {
  thread: Thread
  messages: ChatMessage[]
}
