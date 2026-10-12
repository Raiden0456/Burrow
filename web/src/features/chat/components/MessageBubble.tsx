import { Markdown } from "@/components/Markdown"
import { cn } from "@/lib/utils"
import type { ChatRole } from "../types/chat"

export function MessageBubble({ role, content }: { role: ChatRole; content: string }) {
  if (role === "user") {
    return <div className="ml-auto max-w-[85%] rounded-2xl rounded-br-sm bg-primary px-4 py-2 text-sm whitespace-pre-wrap text-primary-foreground">{content}</div>
  }
  return (
    <div className={cn("max-w-full border-l-2 border-hue-purple/50 pl-4")}>
      <Markdown className="prose-sm">{content}</Markdown>
    </div>
  )
}
